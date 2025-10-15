package etcd

import (
	"fmt"
	"log"
	"testing"
	"time"
)

var lock *DistributedLock

func init() {
	lock = NewDistributedLock([]string{"127.0.0.1:2379"})
}

func TestTryLock(t *testing.T) {
	session, mutex, err := lock.TryLock("/my-lock", 5*time.Second)
	if err != nil {
		log.Fatal("获取锁失败:", err)
	}
	fmt.Println("成功获取锁")

	// 执行业务逻辑
	fmt.Println("执行关键业务逻辑...")
	time.Sleep(3 * time.Second)

	// 释放锁
	err = lock.Unlock(session, mutex)
	if err != nil {
		log.Fatal("释放锁失败:", err)
	}
	fmt.Println("成功释放锁")
}
