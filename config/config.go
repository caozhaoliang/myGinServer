package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var once sync.Once
var config *Config

func GetConfig() *Config {
	return config
}
func InitConfig(configYaml string) *Config {
	once.Do(func() {
		// 读取yaml文件生成配置实例
		config = new(Config)
		data, err := os.ReadFile(configYaml)
		if err != nil {
			fmt.Printf("读取文件失败: %v\n", err)
			panic(err)
		}

		// 2. 解析 YAML 到结构体
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			fmt.Printf("解析 YAML 失败: %v\n", err)
			panic(err)
		}
	})
	return config
}

type Config struct {
	DBConfig     DBConfig           `yaml:"db"`
	ObjectConfig ObjectServerConfig `yaml:"object"`
	Cache        CacheConfig        `yaml:"cache"`
	SSH          SSHConfig          `yaml:"ssh"`
}

// SSHConfig 远程执行（shell 节点）的 SSH 连接配置。
// 认证方式二选一：password 或 rsa_private_key_path / rsa_private_key（内联 PEM）。
type SSHConfig struct {
	Host              string `yaml:"host"`                // 形如 127.0.0.1:22
	User              string `yaml:"user"`                // 登录用户
	Password          string `yaml:"password"`            // 密码认证（可选）
	RsaPrivateKeyPath string `yaml:"rsa_private_key_path"` // 私钥文件路径（与内联私钥二选一）
	RsaPrivateKey     string `yaml:"rsa_private_key"`     // 内联 PEM 私钥（与私钥路径二选一）
	Passphrase        string `yaml:"passphrase"`          // 私钥口令（无口令时留空）
	ShellType         string `yaml:"shell_type"`          // 远程解释器：sh / bash（默认 bash）
	TimeoutSeconds    int    `yaml:"timeout_seconds"`     // 连接超时秒数（默认 5）
}

type DBConfig struct {
	DbHost     string `yaml:"db_host"`
	DbPort     int    `yaml:"db_port"`
	DbUser     string `yaml:"db_user"`
	DbPassword string `yaml:"db_password"`
	DbName     string `yaml:"db_name"`
}

type ObjectServerConfig struct {
	Endpoint     string `yaml:"endpoint"`
	AccessID     string `yaml:"access_id"`
	AccessSecret string `yaml:"access_secret"`
	BucketName   string `yaml:"bucket_name"`
	RootPath     string `yaml:"root_path"`
}

type CacheConfig struct {
	RedisHost string `yaml:"redis_host"`
	RedisPort string `yaml:"redis_port"`
	Password  string `yaml:"password"`
	PoolSize  int    `yaml:"pool_size"`
	DB        int    `yaml:"db"`
}
