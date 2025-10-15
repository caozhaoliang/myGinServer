package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ConfigManager struct {
	cli    *clientv3.Client
	config map[string]interface{}
}

func NewConfigManager(endpoints []string) *ConfigManager {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &ConfigManager{
		cli:    cli,
		config: make(map[string]interface{}),
	}
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig(key string, config interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = cm.cli.Put(context.Background(), key, string(configJSON))
	return err
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig(key string, config interface{}) error {
	resp, err := cm.cli.Get(context.Background(), key)
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("config not found")
	}

	return json.Unmarshal(resp.Kvs[0].Value, config)
}

// WatchConfig 监听配置变化
func (cm *ConfigManager) WatchConfig(key string, callback func(config interface{})) {
	rch := cm.cli.Watch(context.Background(), key)

	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == clientv3.EventTypePut {
				var config interface{}
				if err := json.Unmarshal(ev.Kv.Value, &config); err == nil {
					callback(config)
				}
			}
		}
	}
}
