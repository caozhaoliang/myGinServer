package trace

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// 示例：一个处理用户请求的函数
func handleUserRequest(ctx context.Context, userID string) error {
	// 获取全局 tracer（指定当前模块名）
	tracer := otel.Tracer("user-handler")

	// 创建一个 Span（作为当前上下文的子 Span）
	ctx, span := tracer.Start(ctx, "handleUserRequest")
	defer span.End() // 函数结束时关闭 Span（自动记录耗时）

	// 添加标签（Key-Value 形式，用于筛选和分析）
	span.SetAttributes(
		attribute.String("user.id", userID),    // 用户 ID
		attribute.String("operation", "query"), // 操作类型
	)

	// 模拟数据库查询（子操作）
	if err := queryDatabase(ctx, userID); err != nil {
		// 记录错误日志（会关联到当前 Span）
		span.RecordError(err)
		span.SetStatus(codes.Error, "数据库查询失败")
		return err
	}

	return nil
}

// 数据库查询（子 Span）
func queryDatabase(ctx context.Context, userID string) error {
	_, span := otel.Tracer("db-client").Start(ctx, "queryDatabase")
	defer span.End()

	span.SetAttributes(attribute.String("db.table", "users"))
	// 模拟查询耗时
	time.Sleep(10 * time.Millisecond)
	return nil
}
func httpHandler(w http.ResponseWriter, r *http.Request) {
	// 从请求 Header 中提取追踪上下文
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	// 基于提取的上下文创建 Span
	_, span := otel.Tracer("http-server").Start(ctx, "httpHandler")
	defer span.End()

	// 处理请求...
	handleUserRequest(ctx, r.URL.Query().Get("user_id"))
}
func TestTraceInit(t *testing.T) {
	tracer, err := initTracer("http://localhost:14268/api/traces")
	if err != nil {
		t.Error(err)
	}
	defer tracer(context.TODO())
	// 启动 HTTP 服务
	http.HandleFunc("/user", httpHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
