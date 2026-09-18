// Package migrate 提供类似 DataX 的 MySQL → MySQL 数据迁移能力：
// 流式读取源表 → 按批（BatchSize）事务写入目标表，支持 insert / replace / upsert
// 三种写入模式、可选 where 过滤、目标表自动建表，以及按数字主键区间切分的并发通道。
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"myGinServer/api/response"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

const (
	StatusRunning  = "running"
	StatusSuccess  = "success"
	StatusFailed   = "failed"
	StatusCanceled = "canceled"

	DefaultBatchSize = 2000
)

// Endpoint MySQL 连接信息（源 / 目标）
type Endpoint struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// Config 一次迁移任务的完整配置
type Config struct {
	Source               Endpoint
	Target               Endpoint
	Table                string   // 源表名
	TargetTable          string   // 目标表名（空 = 与源同名）
	Columns              []string // 目标列（空 = 源表全列）
	Where                string
	BatchSize            int
	Mode                 string // insert / replace / upsert
	CreateTableIfMissing bool
	CreateDDL            string // 自定义建表 DDL（优先于源表 SHOW CREATE TABLE）
	Channels             int
}

// targetTable 目标表名（缺省回退到源表名）
func (c Config) targetTable() string {
	if c.TargetTable != "" {
		return c.TargetTable
	}
	return c.Table
}

var (
	ErrTaskExists = errors.New("任务已存在（run_id 重复）")
	ErrTaskNotFound = errors.New("任务不存在")
	ErrTaskFinished = errors.New("任务已结束")
)

// Migrator 迁移任务管理器（内存态：进程重启后任务记录丢失，如需持久化可后续落库）
type Migrator struct {
	mu    sync.Mutex
	tasks map[string]*Task
}

func NewMigrator() *Migrator {
	return &Migrator{tasks: make(map[string]*Task)}
}

// Task 单个迁移任务的运行时状态
type Task struct {
	cfg Config

	mu          sync.Mutex
	status      string
	message     string
	startedAt   time.Time
	finishedAt  time.Time
	rowsRead    int64
	rowsWritten int64
	batches     int64
	cancel      context.CancelFunc
}

// Run 提交并异步执行迁移任务。run_id 幂等：运行中的任务重复提交返回 ErrTaskExists；
// 已结束（success/failed/canceled）的任务允许用同一 run_id 重新提交覆盖。
func (m *Migrator) Run(runId string, cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tasks[runId]; ok {
		t.mu.Lock()
		running := t.status == StatusRunning
		t.mu.Unlock()
		if running {
			return ErrTaskExists
		}
	}
	task := &Task{
		cfg:       cfg,
		status:    StatusRunning,
		startedAt: time.Now(),
	}
	m.tasks[runId] = task
	go m.execute(task)
	return nil
}

// Stop 取消运行中的任务（等待执行协程检测到 ctx 取消后置为 canceled）。
func (m *Migrator) Stop(runId string) error {
	m.mu.Lock()
	task, ok := m.tasks[runId]
	m.mu.Unlock()
	if !ok {
		return ErrTaskNotFound
	}
	task.mu.Lock()
	cancel := task.cancel
	finished := task.status != StatusRunning
	task.mu.Unlock()
	if finished {
		return ErrTaskFinished
	}
	if cancel != nil {
		cancel()
	}
	return nil
}

// Status 返回任务快照；第二返回值为 false 表示任务不存在。
func (m *Migrator) Status(runId string) (response.MigrateStatus, bool) {
	m.mu.Lock()
	task, ok := m.tasks[runId]
	m.mu.Unlock()
	if !ok {
		return response.MigrateStatus{}, false
	}
	task.mu.Lock()
	defer task.mu.Unlock()
	return response.MigrateStatus{
		RunId:       runId,
		Status:      task.status,
		RowsRead:    atomic.LoadInt64(&task.rowsRead),
		RowsWritten: atomic.LoadInt64(&task.rowsWritten),
		Batches:     atomic.LoadInt64(&task.batches),
		Message:     task.message,
		StartedAt:   task.startedAt,
		FinishedAt:  task.finishedAt,
	}, true
}

func (m *Migrator) execute(task *Task) {
	ctx, cancel := context.WithCancel(context.Background())
	task.mu.Lock()
	task.cancel = cancel
	task.mu.Unlock()
	defer cancel()

	status, message := StatusSuccess, "迁移完成"
	defer func() {
		if r := recover(); r != nil {
			status, message = StatusFailed, fmt.Sprintf("任务异常终止: %v", r)
		}
		task.mu.Lock()
		task.status = status
		task.message = message
		task.finishedAt = time.Now()
		task.mu.Unlock()
	}()

	if err := m.run(ctx, task); err != nil {
		if ctx.Err() != nil {
			status, message = StatusCanceled, "任务已取消"
		} else {
			status, message = StatusFailed, err.Error()
		}
	}
}

func (m *Migrator) run(ctx context.Context, task *Task) error {
	cfg := task.cfg
	src, err := openDB(cfg.Source)
	if err != nil {
		return errors.Wrap(err, "连接源库失败")
	}
	defer src.Close()
	tgt, err := openDB(cfg.Target)
	if err != nil {
		return errors.Wrap(err, "连接目标库失败")
	}
	defer tgt.Close()

	// 1. 列清单：未指定时按源表列顺序取全列
	cols := cfg.Columns
	if len(cols) == 0 {
		cols, err = resolveColumns(ctx, src, cfg.Source.Database, cfg.Table)
		if err != nil {
			return errors.Wrap(err, "解析源表列失败")
		}
		if len(cols) == 0 {
			return fmt.Errorf("源表 %s 不存在或没有列", cfg.Table)
		}
	}

	// 2. 可选：目标表不存在时用 SHOW CREATE TABLE 建表
	if cfg.CreateTableIfMissing {
		if err := ensureTargetTable(ctx, src, tgt, cfg); err != nil {
			return errors.Wrap(err, "创建目标表失败")
		}
	}
	batch := cfg.BatchSize
	if batch <= 0 {
		batch = DefaultBatchSize
	}
	mode := cfg.Mode
	if mode == "" {
		mode = "insert"
	}

	// 3. 并发通道：仅当 channels>1 且源表有数字主键时按主键区间并行，否则串行
	channels := cfg.Channels
	if channels > 1 {
		pkCol, dataType, err := probePrimaryKey(ctx, src, cfg.Source.Database, cfg.Table)
		if err == nil && pkCol != "" && isNumericPKType(dataType) {
			return m.runParallel(ctx, task, src, tgt, cols, pkCol, cfg, batch, channels)
		}
		task.mu.Lock()
		task.message = "源表无数字主键，channels 已回退为串行执行"
		task.mu.Unlock()
	}
	return runSerial(ctx, task, src, tgt, cols, cfg, batch, "", nil)
}

// runSerial 串行执行一段迁移：SELECT（可选 pk 区间）→ 分批事务写入。
// pkCol 非空时，pkCond 形如 "`id` BETWEEN ? AND ?"，与 where 条件 AND 拼接。
func runSerial(ctx context.Context, task *Task, src, tgt *sql.DB, cols []string, cfg Config, batch int, pkCol string, rng *pkRange) error {
	where := strings.TrimSpace(cfg.Where)
	args := []interface{}{}
	if rng != nil {
		pkCond := fmt.Sprintf("%s BETWEEN ? AND ?", QuoteIdent(pkCol))
		if where != "" {
			where = "(" + where + ") AND " + pkCond
		} else {
			where = pkCond
		}
		args = append(args, rng.Lo, rng.Hi)
	}
	query, err := buildSelect(cfg.Table, cols, where)
	if err != nil {
		return err
	}

	rows, err := src.QueryContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "查询源表失败")
	}
	defer rows.Close()

	buf := make([][]interface{}, 0, batch)
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		vals, err := scanRow(rows, len(cols))
		if err != nil {
			return err
		}
		buf = append(buf, vals)
		atomic.AddInt64(&task.rowsRead, 1)
		if len(buf) >= batch {
			if err := writeBatch(ctx, task, tgt, cfg, cols, buf); err != nil {
				return err
			}
			buf = buf[:0]
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "读取源表失败")
	}
	if len(buf) > 0 {
		if err := writeBatch(ctx, task, tgt, cfg, cols, buf); err != nil {
			return err
		}
	}
	return nil
}

// runParallel 按数字主键切分区间，每区间一个 goroutine 并行迁移。
// 任一区间失败即取消全部并返回错误。
func (m *Migrator) runParallel(ctx context.Context, task *Task, src, tgt *sql.DB, cols []string, pkCol string, cfg Config, batch, channels int) error {
	where := strings.TrimSpace(cfg.Where)
	minMaxSQL := "SELECT MIN(" + QuoteIdent(pkCol) + "), MAX(" + QuoteIdent(pkCol) + ") FROM " + QuoteIdent(cfg.Table)
	if where != "" {
		minMaxSQL += " WHERE " + where
	}
	var minV, maxV sql.NullInt64
	if err := src.QueryRowContext(ctx, minMaxSQL).Scan(&minV, &maxV); err != nil {
		return errors.Wrap(err, "探测主键范围失败")
	}
	if !minV.Valid || !maxV.Valid {
		return nil // 空表：无数据可迁
	}

	ranges := splitPKRanges(minV.Int64, maxV.Int64, channels)
	errCh := make(chan error, len(ranges))
	var wg sync.WaitGroup
	for _, r := range ranges {
		wg.Add(1)
		go func(r pkRange) {
			defer wg.Done()
			if err := runSerial(ctx, task, src, tgt, cols, cfg, batch, pkCol, &r); err != nil {
				errCh <- err
			}
		}(r)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}

// writeBatch 在单个事务内批量写入一批行（写入目标表 cfg.targetTable()）。
func writeBatch(ctx context.Context, task *Task, tgt *sql.DB, cfg Config, cols []string, rows [][]interface{}) error {
	stmt, err := buildInsert(cfg.targetTable(), cols, len(rows), cfg.Mode)
	if err != nil {
		return err
	}
	args := make([]interface{}, 0, len(cols)*len(rows))
	for _, r := range rows {
		args = append(args, r...)
	}
	tx, err := tgt.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "开启目标事务失败")
	}
	if _, err := tx.ExecContext(ctx, stmt, args...); err != nil {
		_ = tx.Rollback()
		return errors.Wrapf(err, "写入目标表失败（本批 %d 行）", len(rows))
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "提交目标事务失败")
	}
	atomic.AddInt64(&task.rowsWritten, int64(len(rows)))
	atomic.AddInt64(&task.batches, 1)
	return nil
}

func scanRow(rows *sql.Rows, n int) ([]interface{}, error) {
	vals := make([]interface{}, n)
	dest := make([]interface{}, n)
	for i := range dest {
		dest[i] = &vals[i]
	}
	if err := rows.Scan(dest...); err != nil {
		return nil, errors.Wrap(err, "扫描源表行失败")
	}
	return vals, nil
}

// openDB 建立 MySQL 连接（parseTime=true 保证 DATETIME 读为 time.Time）。
func openDB(e Endpoint) (*sql.DB, error) {
	c := mysql.NewConfig()
	c.User = e.User
	c.Passwd = e.Password
	c.Net = "tcp"
	c.Addr = fmt.Sprintf("%s:%d", e.Host, e.Port)
	c.DBName = e.Database
	c.Timeout = 5 * time.Second
	c.ReadTimeout = 5 * time.Minute
	c.Params = map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "true",
		"loc":       "Local",
	}
	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// resolveColumns 按 ORDINAL_POSITION 取源表全部列名。
func resolveColumns(ctx context.Context, src *sql.DB, database, table string) ([]string, error) {
	rows, err := src.QueryContext(ctx,
		"SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION",
		database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

// probePrimaryKey 探测单列主键的列名与数据类型（复合主键返回空，不做并发切分）。
func probePrimaryKey(ctx context.Context, src *sql.DB, database, table string) (string, string, error) {
	rows, err := src.QueryContext(ctx, pkProbeSQL, database, table)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()
	var pkCol, dataType string
	count := 0
	for rows.Next() {
		if err := rows.Scan(&pkCol, &dataType); err != nil {
			return "", "", err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return "", "", err
	}
	if count != 1 {
		return "", "", nil // 无主键或复合主键
	}
	return pkCol, dataType, nil
}

// ensureTargetTable 目标表不存在时，用源表的 SHOW CREATE TABLE 原样建表（目标表名）。
func ensureTargetTable(ctx context.Context, src, tgt *sql.DB, cfg Config) error {
	targetTable := cfg.targetTable()
	var cnt int
	err := tgt.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?",
		cfg.Target.Database, targetTable).Scan(&cnt)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	// 自定义 DDL 优先：节点 content 里配置了 create_ddl（如收集目标建表语句）时直接使用
	if cfg.CreateDDL != "" {
		if _, err := tgt.ExecContext(ctx, cfg.CreateDDL); err != nil {
			return errors.Wrap(err, "执行自定义建表 DDL 失败")
		}
		return nil
	}
	var tableName, ddl string
	if err := src.QueryRowContext(ctx, "SHOW CREATE TABLE "+QuoteIdent(cfg.Table)).Scan(&tableName, &ddl); err != nil {
		return errors.Wrap(err, "读取源表 DDL 失败")
	}
	if targetTable != cfg.Table {
		ddl = strings.ReplaceAll(ddl, QuoteIdent(cfg.Table), QuoteIdent(targetTable))
	}
	if _, err := tgt.ExecContext(ctx, ddl); err != nil {
		return errors.Wrap(err, "执行源表 DDL 失败")
	}
	return nil
}
