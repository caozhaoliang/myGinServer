package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type LeaderElection struct {
	cli      *clientv3.Client
	election *concurrency.Election
}

func NewLeaderElection(endpoints []string, electionKey string) *LeaderElection {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &LeaderElection{
		cli: cli,
	}
}

// Campaign 参与选举
func (le *LeaderElection) Campaign(nodeID string) error {
	session, err := concurrency.NewSession(le.cli, concurrency.WithTTL(10))
	if err != nil {
		return err
	}

	election := concurrency.NewElection(session, "/leader-election/")
	le.election = election

	// 参与选举
	err = election.Campaign(context.Background(), nodeID)
	if err != nil {
		return err
	}

	fmt.Printf("节点 %s 成为Leader\n", nodeID)

	// 作为Leader执行业务逻辑
	go le.leaderWork(session)

	return nil
}

// leaderWork Leader的工作
func (le *LeaderElection) leaderWork(session *concurrency.Session) {
	defer session.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Leader正在工作...")
		case <-session.Done():
			fmt.Println("不再是Leader")
			return
		}
	}
}

// Observe 监听Leader变化
func (le *LeaderElection) Observe() {
	session, err := concurrency.NewSession(le.cli)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	election := concurrency.NewElection(session, "/leader-election/")
	ch := election.Observe(context.Background())

	for {
		select {
		case resp := <-ch:
			if len(resp.Kvs) > 0 {
				fmt.Printf("当前Leader: %s\n", resp.Kvs[0].Value)
			}
		}
	}
}
