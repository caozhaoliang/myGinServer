package etcd

import (
	"context"
	"testing"
	"time"
)

var discovery *ServiceDiscovery

func init() {

	discovery = NewServiceDiscovery([]string{"http://127.0.0.1:2379"})
}
func TestNewServiceDiscovery(t *testing.T) {
	go func() {
		chanWatch, err := discovery.WatchServices(context.Background(), "test")
		if err != nil {
			t.Error(err)
		}
		for {
			select {
			case service := <-chanWatch:
				t.Log(service)
			}
		}
	}()
	err := discovery.RegisterService("test", "127.0.0.1:8080", 5)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
	err = discovery.RegisterService("test", "127.0.0.2:8080", 5)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
}
