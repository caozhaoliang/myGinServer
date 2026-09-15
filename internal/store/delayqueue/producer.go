package delayqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Producer struct {
	db *gorm.DB
}

func NewProducer(db *gorm.DB) *Producer {
	return &Producer{db: db}
}

const (
	StatusPending = 0 // 待处理
	StatusRunning = 1 // 处理中
	StatusSuccess = 2 // 成功
	StatusFailed  = 3 // 失败
)

type DelayMessage struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	BizID       string    `gorm:"column:biz_id;size:64;not null"`
	Topic       string    `gorm:"column:topic;size:64;not null"`
	ExecuteTime time.Time `gorm:"column:execute_time;type:datetime(3);not null"`
	Status      int       `gorm:"column:status;not null;default:0"`
	Payload     string    `gorm:"column:payload;type:json"`
	RetryCount  int       `gorm:"column:retry_count;not null;default:0"`
	MaxRetry    int       `gorm:"column:max_retry;not null;default:3"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (DelayMessage) TableName() string { return "delay_queue" }

// Publish 投递一条延时消息，delay 表示延迟多久执行
func (p *Producer) Publish(ctx context.Context, topic, bizID string, payload interface{}, delay time.Duration) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg := DelayMessage{
		BizID:       bizID,
		Topic:       topic,
		ExecuteTime: time.Now().Add(delay),
		Status:      StatusPending,
		Payload:     string(body),
		MaxRetry:    3,
	}
	return p.db.WithContext(ctx).Create(&msg).Error
}

func (p *Producer) PublishTx(tx *gorm.DB, ctx context.Context, topic, bizID string, payload interface{}, delay time.Duration) error {
	body, _ := json.Marshal(payload)
	msg := DelayMessage{
		BizID:       bizID,
		Topic:       topic,
		ExecuteTime: time.Now().Add(delay),
		Status:      StatusPending,
		Payload:     string(body),
		MaxRetry:    3,
	}
	return tx.WithContext(ctx).Create(&msg).Error
}
