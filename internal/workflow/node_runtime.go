package workflow

import (
	"context"
	"fmt"
	"myGinServer/internal/store/delayqueue"
	"myGinServer/internal/store/dispatch"
	mdispatch "myGinServer/models/dispatch"
	"time"

	"github.com/pkg/errors"
)

type NodeRuntime struct {
	saas     dispatch.StoreIface
	consumer *delayqueue.Consumer
	producer *delayqueue.Producer
}

func NewNodeRuntime(saas dispatch.StoreIface,
	producer *delayqueue.Producer,
	consumer *delayqueue.Consumer) *NodeRuntime {
	return &NodeRuntime{saas: saas, producer: producer, consumer: consumer}
}

func (c *NodeRuntime) Dispatch(topic string) error {
	c.consumer.Register(topic, func(ctx context.Context, msg *delayqueue.DelayMessage) error {
		return c.deal(ctx, msg.Payload)
	})

	return nil
}

func (c *NodeRuntime) preLoad(ctx context.Context, instanceId string) (*RuntimeCtx, error) {
	instance, err := c.saas.GetInstance(ctx, "", instanceId)
	if err != nil {
		return nil, err
	}
	node, err := c.saas.GetNode(ctx, "", instance.NodeId)
	if err != nil {
		return nil, err
	}
	return &RuntimeCtx{
		node:     *node,
		instance: *instance,
	}, nil
}

func (c *NodeRuntime) deal(ctx context.Context, payload string) error {
	//  根据payload中存储的实例ID获取实例、节点的信息，判断是否存在同节点不同实例正在运行，如果是则跳过。
	count, err := c.saas.InstanceRunningCount(ctx, "", payload)
	if err != nil {
		return err
	}
	if count > 0 {
		_ = c.saas.UpdateInstance(ctx, "", payload, string(mdispatch.InstanceStatusAbort))
		return nil
	}
	reqCtx, err := c.preLoad(ctx, payload)
	if err != nil {
		return err
	}
	// 根据节点类型，组装执行器，执行实例任务。

	// 触发下游实例。
	err = c.finishInstance(ctx, mdispatch.InstanceStatusSuccess, reqCtx)
	return err
}

func (c *NodeRuntime) finishInstance(ctx context.Context, status mdispatch.InstanceStatus, reqCtx *RuntimeCtx) error {
	err := c.saas.UpdateInstance(ctx, "", reqCtx.instance.Id, string(mdispatch.InstanceStatusSuccess))
	if err != nil {
		return err
	}

	if status == mdispatch.InstanceStatusSuccess {
		err = c.tryNextInstances(ctx, reqCtx)
	}

	return err
}
func (c *NodeRuntime) tryNextInstances(ctx context.Context, reqCtx *RuntimeCtx) error {
	rows, err := c.saas.TryNextRows(ctx, "", reqCtx.instance.Id)
	if err != nil {
		return err
	}
	successBehindIds := getAlreadyInfo(rows)
	for k, v := range successBehindIds {
		// 同 dispatch.go：Second() 是分钟内的秒序号，相减得不到时间差；
		// 直接用「距预期执行时间的剩余间隔」，预期时间已过时返回负值，消费端会立即取走。
		err = c.producer.Publish(ctx, mdispatch.NodeInstanceTopic, k, k, time.Until(v))
		if err != nil {
			return errors.Wrap(err, "投递延时队列失败")
		}
		fmt.Printf("投递下游成功：%s", k)
	}

	return err
}

func getAlreadyInfo(rows []mdispatch.TryNextRow) map[string]time.Time {
	// 1. 按下游分组
	type downstreamAgg struct {
		executeTime time.Time
		total       int // 该下游的上游总数
		success     int // 上游中状态为 Success 的数量
	}
	agg := make(map[string]*downstreamAgg)

	for _, r := range rows {
		a, ok := agg[r.DownstreamID]
		if !ok {
			a = &downstreamAgg{executeTime: r.ExecuteTime.Time}
			agg[r.DownstreamID] = a
		}
		a.total++
		if r.UpstreamStat == string(mdispatch.InstanceStatusSuccess) {
			a.success++
		}
	}

	// 2. 判断：总数 > 0 且 全部成功
	ready := make(map[string]time.Time, len(agg))
	for downstreamID, a := range agg {
		if a.total > 0 && a.total == a.success {
			ready[downstreamID] = a.executeTime
		}
	}
	return ready
}
