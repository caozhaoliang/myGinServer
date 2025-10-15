package etcd

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type TaskScheduler struct {
	nodeID     string
	etcdClient *clientv3.Client
	election   *concurrency.Election
	session    *concurrency.Session
	isLeader   bool
	leaderID   string
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewTaskScheduler(endpoints []string, nodeID string) *TaskScheduler {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &TaskScheduler{
		nodeID:     nodeID,
		etcdClient: cli,
		stopCh:     make(chan struct{}),
	}
}

func (ts *TaskScheduler) Start() error {
	// 创建session
	session, err := concurrency.NewSession(ts.etcdClient, concurrency.WithTTL(10))
	if err != nil {
		return err
	}
	ts.session = session

	// 创建选举
	election := concurrency.NewElection(session, "/task-scheduler/leader/")
	ts.election = election

	// 启动Leader选举
	ts.wg.Add(2)
	go ts.campaignForLeadership()
	go ts.observeLeadership()

	log.Printf("TaskScheduler %s started, participating in leader election", ts.nodeID)
	return nil
}

func (ts *TaskScheduler) Stop() {
	close(ts.stopCh)
	if ts.session != nil {
		ts.session.Close()
	}
	ts.etcdClient.Close()
	ts.wg.Wait()
	log.Printf("TaskScheduler %s stopped", ts.nodeID)
}

// 参与Leader选举
func (ts *TaskScheduler) campaignForLeadership() {
	defer ts.wg.Done()

	for {
		select {
		case <-ts.stopCh:
			return
		default:
			// 尝试成为Leader
			log.Printf("Node %s is campaigning for leadership...", ts.nodeID)
			err := ts.election.Campaign(context.Background(), ts.nodeID)
			if err != nil {
				log.Printf("Node %s campaign failed: %v", ts.nodeID, err)
				time.Sleep(2 * time.Second)
				continue
			}

			// 成为Leader
			ts.isLeader = true
			log.Printf("🎉 Node %s became the LEADER", ts.nodeID)

			// 作为Leader执行任务
			ts.leaderWork()

			// Leader工作结束，重新参与选举
			ts.isLeader = false
		}
	}
}

// Leader的工作内容
func (ts *TaskScheduler) leaderWork() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ts.stopCh:
			return
		case <-ts.session.Done():
			log.Printf("Node %s session expired, no longer leader", ts.nodeID)
			return
		case <-ticker.C:
			// 只有Leader执行的任务
			ts.executeLeaderTasks()
		}
	}
}

// 执行Leader专属任务
func (ts *TaskScheduler) executeLeaderTasks() {
	if !ts.isLeader {
		return
	}

	tasks := []string{
		"分配定时任务到工作节点",
		"清理过期的任务记录",
		"重新平衡工作负载",
		"检查节点健康状态",
		"更新任务调度计划",
	}

	task := tasks[rand.Intn(len(tasks))]
	log.Printf("🏆 Leader %s is executing: %s", ts.nodeID, task)

	// 模拟任务执行时间
	time.Sleep(500 * time.Millisecond)
}

// 监听Leader变化
func (ts *TaskScheduler) observeLeadership() {
	defer ts.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ts.stopCh
		cancel()
	}()

	leaderCh := ts.election.Observe(ctx)

	for {
		select {
		case <-ts.stopCh:
			return
		case resp, ok := <-leaderCh:
			if !ok {
				log.Printf("Node %s leadership observation channel closed", ts.nodeID)
				return
			}

			if len(resp.Kvs) > 0 {
				newLeaderID := string(resp.Kvs[0].Value)
				ts.leaderID = newLeaderID

				if newLeaderID == ts.nodeID {
					log.Printf("✅ Node %s confirmed as current leader", ts.nodeID)
				} else {
					log.Printf("📢 Node %s acknowledges new leader: %s", ts.nodeID, newLeaderID)
				}
			} else {
				log.Printf("⚠️ No leader elected currently")
				ts.leaderID = ""
			}
		}
	}
}

// 获取当前状态信息
func (ts *TaskScheduler) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"node_id":   ts.nodeID,
		"is_leader": ts.isLeader,
		"leader_id": ts.leaderID,
		"active":    ts.session != nil,
	}
}
