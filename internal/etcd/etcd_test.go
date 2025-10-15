package etcd

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var client *clientv3.Client

func init() {
	// 配置ETCD客户端
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // ETCD节点地址
		DialTimeout: 5 * time.Second,            // 连接超时时间
	}
	var err error
	// 建立连接
	client, err = clientv3.New(config)
	if err != nil {
		log.Fatal(err)
	}
	// defer client.Close()

	fmt.Println("连接ETCD成功")
}

func TestPutGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PUT操作
	_, err := client.Put(ctx, "myFirstKey", "zLiang.cao")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PUT操作成功")
	resp, err := client.Get(ctx, "myFirstKey")
	if err != nil {
		log.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		fmt.Printf("key:%s, value:%s\n", kv.Key, kv.Value)
	}
}

// 使用etcd进行监听键的变化
func TestWatch(t *testing.T) {
	go func() {
		time.Sleep(5 * time.Second)
		client.Put(context.Background(), "myFirstKey", "zLiang.cao11")
		time.Sleep(1 * time.Second)
		client.Put(context.Background(), "myFirstKey", "zLiang.cao22")
	}()
	rch := client.Watch(context.Background(), "myFirstKey")
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

// 测试租约功能
func TestLease(t *testing.T) {
	// 创建租约
	lease := clientv3.NewLease(client)
	// 设置租约时间为10秒
	leaseResp, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		log.Fatal(err)
	}

	leaseID := leaseResp.ID

	// 使用租约PUT一个键
	_, err = client.Put(context.TODO(), "lease_key", "lease_value", clientv3.WithLease(leaseID))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PUT带租约的键成功，10秒后自动删除")

	// 持续检查键是否存在
	for {
		resp, err := client.Get(context.TODO(), "lease_key")
		if err != nil {
			log.Fatal(err)
		}
		if len(resp.Kvs) == 0 {
			fmt.Println("键已过期")
			break
		}
		fmt.Println("键还存在")
		time.Sleep(2 * time.Second)
	}
}

// 测试分布式锁功能
func TestMutexLock(t *testing.T) {
	// 创建两个会话，模拟两个客户端
	s1, err := concurrency.NewSession(client, concurrency.WithTTL(10))
	if err != nil {
		log.Fatal(err)
	}
	defer s1.Close()
	m1 := concurrency.NewMutex(s1, "/my-lock/")

	s2, err := concurrency.NewSession(client, concurrency.WithTTL(10))
	if err != nil {
		log.Fatal(err)
	}
	defer s2.Close()
	m2 := concurrency.NewMutex(s2, "/my-lock/")

	// 客户端1获取锁
	if err := m1.Lock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("客户端1获取锁")

	// 客户端2尝试获取锁
	go func() {
		if err := m2.Lock(context.TODO()); err != nil {
			log.Fatal(err)
		}
		fmt.Println("客户端2获取锁")
	}()

	// 客户端1持有锁一段时间
	time.Sleep(3 * time.Second)

	// 客户端1释放锁
	if err := m1.Unlock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("客户端1释放锁")

	// 等待客户端2获取锁
	time.Sleep(3 * time.Second)
}
