package etcd

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type DistributedLock struct {
	cli *clientv3.Client
}

func NewDistributedLock(endpoints []string) *DistributedLock {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &DistributedLock{cli: cli}
}

// TryLock 尝试获取锁
func (dl *DistributedLock) TryLock(lockKey string, timeout time.Duration) (*concurrency.Session, *concurrency.Mutex, error) {
	// 创建session
	session, err := concurrency.NewSession(dl.cli, concurrency.WithTTL(10))
	if err != nil {
		return nil, nil, err
	}

	// 创建mutex
	mutex := concurrency.NewMutex(session, lockKey)

	// 带超时的获取锁
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = mutex.Lock(ctx)
	if err != nil {
		session.Close()
		return nil, nil, err
	}

	return session, mutex, nil
}

// Unlock 释放锁
func (dl *DistributedLock) Unlock(session *concurrency.Session, mutex *concurrency.Mutex) error {
	defer session.Close()
	return mutex.Unlock(context.Background())
}
