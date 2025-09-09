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
	DBConfig DBConfig `yaml:"db"`
}

type DBConfig struct {
	DbHost     string `yaml:"db_host"`
	DbPort     int    `yaml:"db_port"`
	DbUser     string `yaml:"db_user"`
	DbPassword string `yaml:"db_password"`
	DbName     string `yaml:"db_name"`
}
