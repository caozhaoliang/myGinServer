package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceDiscovery struct {
	cli        *clientv3.Client
	serverList map[string]string
}

func NewServiceDiscovery(endpoints []string) *ServiceDiscovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &ServiceDiscovery{
		cli:        cli,
		serverList: make(map[string]string),
	}
}

// RegisterService 服务注册
func (s *ServiceDiscovery) RegisterService(serviceName, serviceAddr string, ttl int64) error {
	// 创建租约
	resp, err := s.cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	// 服务注册键格式: /services/{serviceName}/{serviceAddr}
	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceAddr)
	_, err = s.cli.Put(context.Background(), key, serviceAddr, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// 保持租约活跃
	ch, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			<-ch
		}
	}()

	return nil
}

// DiscoverServices 服务发现
func (s *ServiceDiscovery) DiscoverServices(serviceName string) ([]string, error) {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0)
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}

	return addrs, nil
}

type Service struct {
	Name string
	Addr string
}

// WatchServices 监听服务变化
func (s *ServiceDiscovery) WatchServices(ctx context.Context, serviceName string) (chan Service, error) {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	ch := make(chan Service)
	defer close(ch)
	for wresp := range rch {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				fmt.Printf("Service added: %s -> %s\n", ev.Kv.Key, ev.Kv.Value)
				ch <- Service{
					Name: serviceName,
					Addr: string(ev.Kv.Value),
				}
			case clientv3.EventTypeDelete:
				fmt.Printf("Service deleted: %s\n", ev.Kv.Key)
				ch <- Service{
					Name: serviceName,
					Addr: "",
				}
			}
		}
	}
	return ch, nil
}
