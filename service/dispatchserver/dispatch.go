package dispatchserver

import (
	"context"
	"encoding/json"
	"fmt"
	"myGinServer/api/request"
	"myGinServer/api/response"
	"myGinServer/models/dispatch"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	ch = make(chan TestRunEntity, 100)
	mu sync.Mutex
)

const (
	RootNodeId      = "615966d0-af61-11f1-8f44-866b84541a87"
	OdsDatasourceId = "615966d0-af61-11f1-8f44-866b84548888"
)

func init() {
	// todo 读取数据库中的未结束的exec_queue中的数据写入到channel对象中。
}

type TestRunEntity struct {
	Kind     string // 执行类型：sql / shell（空值按 sql 处理）
	Sql      string
	RunId    string
	TenantDb string
}

// SendEntity 把一次测试运行投递到内存队列。
// kind 为执行类型（sql/shell），content 为已渲染的 SQL 文本或 shell 脚本。
// 队列满时立即返回错误，而不是把调用方挂住：ch 容量为 100，且消费端
// （Dispatch 内的单 goroutine）是串行执行任务的，一旦积压，阻塞在这里的
// HTTP 请求既等不到槽位释放，也无法随客户端断开而取消。
func (n *NodeServer) SendEntity(ctx context.Context, tenantDb, runId, kind, content string) error {
	// 先剔除已取消的请求：否则当 ctx 已取消而 channel 恰有空位时，
	// select 会在「发送」与「ctx.Done()」之间随机选一个，导致已取消的请求仍被投递。
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case ch <- TestRunEntity{
		Kind:     kind,
		Sql:      content,
		RunId:    runId,
		TenantDb: tenantDb,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// 注意：此处的 default 是「立即返回错误」，不是阻塞发送，不会挂住调用方。
		return fmt.Errorf("测试运行队列已满（容量 %d），请稍后重试", cap(ch))
	}
}

func (n *NodeServer) Dispatch(ctx context.Context) {
	// cron.New 默认使用空链，任务函数内的 panic 会直接终止整个进程；
	// 显式挂上 Recover，把 panic 限制在单次调度内。
	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))
	// 先启动调度器
	c.Start()
	// 运行时动态添加任务（调度器已在运行，依然生效）
	id, err := c.AddFunc("30 23 *  *  *", func() {
		now := time.Now()
		date := now.Format("2006-01-02")
		_ = n.InstanceCreate(ctx, request.InstanceCreateReq{
			Id:      RootNodeId,
			Project: "",
			BizDate: date,
		})
	})
	if err != nil {
		panic(err)
	}
	go func() {
		defer c.Remove(id)
		for {
			select {
			case <-ctx.Done():
				return
			case entity, ok := <-ch:
				if !ok {
					return
				}
				if entity.Kind == "shell" {
					n.runShell(ctx, entity)
				} else {
					n.runSQL(ctx, entity)
				}
			}
		}
	}()
}

func normalizeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func (n *NodeServer) runSQL(ctx context.Context, entity TestRunEntity) {
	queue, errQuery := n.store.QueryExecQueue(ctx, entity.TenantDb, entity.RunId)
	if errQuery != nil {
		return
	}
	mu.Lock()
	// 判断 queue对象是否已经在运行了或者完成了
	if dispatch.EntityAlreadyRun(queue.Status) {
		mu.Unlock()
		return
	}
	_ = n.store.UpdateExecQueueResp(ctx, entity.TenantDb, entity.RunId, "{}", "running")
	mu.Unlock()
	content, err := n.getDatasourceContent(ctx, "", OdsDatasourceId)
	if err != nil {
		return
	}
	// 打开目标数据库连接
	db, err := openTargetDB(content)
	if err != nil {
		return
	}
	defer db.Close()

	resp := response.TestRunResp{}
	status := "success"
	rows, err := db.QueryContext(ctx, entity.Sql)
	if err != nil {
		resp.Msg = err.Error()
	} else {
		defer rows.Close()
		// 1. 取列名作为 Header
		columns, err := rows.Columns()
		if err != nil {
			resp.Msg = err.Error()
		} else {
			resp.Header = columns

			// 2. 逐行扫描
			var body []map[string]interface{}
			for rows.Next() {
				// 每行准备一组 interface{} 指针
				values := make([]interface{}, len(columns))
				ptrs := make([]interface{}, len(columns))
				for i := range values {
					ptrs[i] = &values[i]
				}

				if err = rows.Scan(ptrs...); err != nil {
					resp.Msg = err.Error()
					return // 或直接 return，看函数签名
				}

				row := make(map[string]interface{}, len(columns))
				for i, col := range columns {
					row[col] = normalizeValue(values[i])
				}
				body = append(body, row)
			}

			// 3. 检查遍历过程中的错误
			if err = rows.Err(); err != nil {
				resp.Msg = err.Error()
			} else {
				resp.Body = body
				resp.Sql = entity.Sql
				resp.Msg = "success"
			}
		}
	}
	if resp.Msg != "success" {
		status = "failed"
	}
	bytes, _ := json.Marshal(resp)
	err = n.store.UpdateExecQueueResp(ctx, entity.TenantDb, entity.RunId, string(bytes), status)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (n *NodeServer) InstanceCreate(ctx context.Context, event request.InstanceCreateReq) error {
	// 根据节点的DAG创建对应实例数据，需要根据schedule的cron进行解析,并将根节点发送到队列中。
	exists, err2 := n.store.InstanceExists(ctx, event.Project, event.BatchId())
	if err2 != nil {
		return err2
	}
	if exists {
		return nil
	}

	lines, err := n.store.ListLines(ctx, event.Project)
	if err != nil {
		return err
	}
	builder := NewInstanceDAGBuilder(event.Id, event.BatchId(), n.store)
	dag, root, err := builder.Build(lines)
	if err != nil {
		return err
	}
	err = dag.Parser(event.BizDate)
	if err != nil {
		return err
	}
	err = n.batchCreateInstance(ctx, root.Id, dag)
	return err
}

func (n *NodeServer) batchCreateInstance(ctx context.Context, rootId string, dag NodeDAG) error {
	// 写入数据库
	depends := dag.BuildInstanceDepend()
	var depend []dispatch.InstanceLine
	for behindID, aheadIds := range depends {
		if len(aheadIds) == 0 {
			continue
		}
		for _, aheadID := range aheadIds {
			depend = append(depend, dispatch.InstanceLine{
				Id:      uuid.New().String(),
				Ahead:   aheadID,
				Behind:  behindID,
				BatchId: dag.BatchId,
			})
		}
	}
	nodes := dag.GetInstanceList()
	err := n.store.BatchCreateInstance(ctx, "", depend, nodes)
	if err != nil {
		return err
	}
	instance, ok := dag.GetInstanceById(rootId)
	if !ok {
		return fmt.Errorf("根节点 %s 在本批次窗口内未生成实例，请检查其 cron 表达式的调度时间", rootId)
	}
	// 延迟时长取「距预期执行时间的剩余间隔」，不能用 Second() 相减：
	// Second() 只是分钟内的秒序号，相减得到的不是时间差。
	// 若预期执行时间已过，time.Until 返回负值，Producer 会落一条过去的 execute_time，消费端立即取走。
	delay := time.Until(instance.ExecuteTime.Time)
	// 写入根实例ID到延时队列。
	err = n.producer.Publish(ctx, dispatch.NodeInstanceTopic, rootId, instance.Id, delay)
	return err
}
