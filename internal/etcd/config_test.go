package etcd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var configManager *ConfigManager

func init() {
	configManager = NewConfigManager([]string{"http://127.0.0.1:2379"})
}

func TestSaveConfig(t *testing.T) {
	err := configManager.SaveConfig("test", "{\"host\":\"127.0.0.1\"}")
	if err != nil {
		t.Error(err)
	}
	var dest string
	err = configManager.GetConfig("test", &dest)
	assert.Nil(t, err)
	fmt.Sprintln(dest)
}

func TestWatchConfig(t *testing.T) {
	go func() {
		configManager.WatchConfig("test", func(newConfig interface{}) {
			fmt.Printf("newConfig:%v\n", newConfig)
		})
	}()

	_ = configManager.SaveConfig("test", "{\"host\":\"127.0.0.2\"}")
	_ = configManager.SaveConfig("test", "{\"host\":\"127.0.0.3\"}")

}
