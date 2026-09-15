package dispatchserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	Sql   string
	RunId string
}

func (n *NodeServer) SendEntity(ctx context.Context, runId, sql string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		ch <- TestRunEntity{
			Sql:   sql,
			RunId: runId,
		}
	}
	return nil
}

func (n *NodeServer) Dispatch(ctx context.Context) {
	c := cron.New()
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
				n.runSQL(ctx, entity)
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
	queue, errQuery := n.store.QueryExecQueue(ctx, "", entity.RunId)
	if errQuery != nil {
		return
	}
	mu.Lock()
	// 判断 queue对象是否已经在运行了或者完成了
	if dispatch.EntityAlreadyRun(queue.Status) {
		mu.Unlock()
		return
	}
	_ = n.store.UpdateExecQueueResp(ctx, "", entity.RunId, "{}", "running")
	mu.Unlock()
	content, err := n.getDatasourceContent(ctx, OdsDatasourceId)
	if err != nil {
		log.Fatalf("获取ods数据源信息失败:%s", err.Error())
		return
	}
	// 打开目标数据库连接
	db, err := openTargetDB(content)
	if err != nil {
		log.Fatalf("打开数据库连接失败：%s", err.Error())
		return
	}
	defer db.Close()

	resp := response.TestRunResp{}
	status := "success"
	rows, err := db.QueryContext(ctx, entity.RunId)
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
				resp.Sql = entity.RunId
				resp.Msg = "success"
			}
		}
	}
	if resp.Msg != "success" {
		status = "failed"
	}
	bytes, _ := json.Marshal(resp)
	err = n.store.UpdateExecQueueResp(ctx, "", entity.RunId, string(bytes), status)
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
	instance := dag.GetInstanceById(rootId)
	delaySec := instance.ExecuteTime.Time.Second() - time.Now().Second()
	// 写入根实例ID到延时队列。
	err = n.producer.Publish(ctx, dispatch.NodeInstanceTopic, rootId, instance.Id, time.Second*time.Duration(delaySec))
	return err
}
