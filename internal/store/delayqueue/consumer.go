package delayqueue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Handler 业务处理函数
type Handler func(ctx context.Context, msg *DelayMessage) error

type Consumer struct {
	db *gorm.DB
	// mu 保护 handlers：Register 通常发生在启动阶段，而 worker goroutine 已可能在读，
	// 二者无锁并发会构成数据竞争。
	mu       sync.RWMutex
	handlers map[string]Handler // topic -> handler
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewConsumer(db *gorm.DB) *Consumer {
	return &Consumer{
		db:       db,
		handlers: make(map[string]Handler),
		interval: time.Second,
		stopCh:   make(chan struct{}),
	}
}

// Register 注册某类任务的处理器
func (c *Consumer) Register(topic string, h Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[topic] = h
}

// lookupHandler 并发安全地取出 topic 对应的处理器
func (c *Consumer) lookupHandler(topic string) (Handler, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	h, ok := c.handlers[topic]
	return h, ok
}

// callHandler 隔离执行 handler，把 panic 收敛成单次调用的 error。
// worker goroutine 处于进程顶层，上方没有 recover，handler 内一旦 panic 会终止整个进程；
// 在这里兜住之后，panic 只会让当前这条消息按失败重试。
func (c *Consumer) callHandler(ctx context.Context, msg *DelayMessage, h Handler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
		}
	}()
	return h(ctx, msg)
}

// processOne 处理一条消息
func (c *Consumer) processOne(ctx context.Context, msg *DelayMessage) {
	handler, ok := c.lookupHandler(msg.Topic)
	if !ok {
		c.markFailed(ctx, msg, fmt.Errorf("no handler for topic %s", msg.Topic))
		return
	}

	// 执行前从 payload 解析参数
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := c.callHandler(execCtx, msg, handler); err != nil {
		c.handleRetry(ctx, msg, err)
		return
	}

	// 成功
	c.db.WithContext(ctx).Model(&DelayMessage{}).
		Where("id = ?", msg.ID).
		Updates(map[string]interface{}{
			"status":     StatusSuccess,
			"updated_at": time.Now(),
		})
}
func (c *Consumer) handleRetry(ctx context.Context, msg *DelayMessage, execErr error) {
	if msg.RetryCount >= msg.MaxRetry {
		c.markFailed(ctx, msg, execErr)
		return
	}

	// 指数退避：1s, 2s, 4s, 8s...
	backoff := time.Duration(1<<uint(msg.RetryCount)) * time.Second
	nextTime := time.Now().Add(backoff)

	c.db.WithContext(ctx).Model(&DelayMessage{}).
		Where("id = ?", msg.ID).
		Updates(map[string]interface{}{
			"status":       StatusPending, // 放回队列
			"retry_count":  msg.RetryCount + 1,
			"execute_time": nextTime,
			"updated_at":   time.Now(),
		})
}

func (c *Consumer) markFailed(ctx context.Context, msg *DelayMessage, execErr error) {
	c.db.WithContext(ctx).Model(&DelayMessage{}).
		Where("id = ?", msg.ID).
		Updates(map[string]interface{}{
			"status":     StatusFailed,
			"updated_at": time.Now(),
		})
}

// fetchOne 抢占一条到期消息，返回 nil 表示当前没有可处理的消息
func (c *Consumer) fetchOne(ctx context.Context) (*DelayMessage, error) {
	var msg DelayMessage
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// FOR UPDATE SKIP LOCKED 抢占
		row := tx.Raw(`
            SELECT * FROM delay_queue
            WHERE status = ? AND execute_time <= NOW(3)
            ORDER BY execute_time ASC
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        `, StatusPending).Scan(&msg)
		if row.Error != nil {
			return row.Error
		}
		if row.RowsAffected == 0 {
			return gorm.ErrRecordNotFound // 没抢到
		}

		// 标记为处理中
		return tx.Model(&DelayMessage{}).
			Where("id = ?", msg.ID).
			Updates(map[string]interface{}{
				"status":     StatusRunning,
				"updated_at": time.Now(),
			}).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// drain 一次把当前所有到期的消息处理完
func (c *Consumer) drain(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.fetchOne(ctx)
		if err != nil {
			return
		}
		if msg == nil {
			return // 没有可处理的
		}
		c.processOne(ctx, msg)
	}
}
func (c *Consumer) StartWorkers(ctx context.Context, n int) {
	for i := 0; i < n; i++ {
		c.wg.Add(1)
		go func(id int) {
			defer c.wg.Done()
			ticker := time.NewTicker(c.interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-c.stopCh:
					return
				case <-ticker.C:
					c.drain(ctx)
				}
			}
		}(i)
	}
}
