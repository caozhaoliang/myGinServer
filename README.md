# MyGinServer

一个基于 **Gin** 的多租户数据管道平台后端（Go 1.24）。以 **画布节点 + DAG 调度** 为核心，支持采集（Collect）、同步（Sync）、SQL、Shell 四类执行节点；内置 **类 DataX 的 MySQL → MySQL 数据迁移** 能力；租户隔离采用 **分库（Database-per-Tenant）** 方案。

## 项目简介

| 能力 | 说明 |
|---|---|
| 多租户 | 租户级分库隔离，请求经 `X-Tenant-Id` 路由到对应租户库 |
| 数据管道 | 可视化画布节点（采集/同步/SQL/Shell）+ 连线 DAG + 实例调度（cron + 延迟队列） |
| 数据迁移 | 类 DataX 的 MySQL → MySQL 表级迁移（全量/增量/覆盖/更新/并发分片/自动建表），节点驱动 |
| 认证授权 | JWT 登录、bcrypt 密码、admin 角色权限中间件 |
| 对象存储 | MinIO 预签名上传（可选依赖） |
| 可观测 | Prometheus metrics 中间件、pprof、Swagger 文档、结构化日志 |

## 技术栈

- **Web 框架**：`gin-gonic/gin` v1.10 + `gin-contrib/gzip` + `gin-contrib/pprof` + `gin-jwt/v2`
- **数据库**：MySQL（`go-sql-driver/mysql` + `jmoiron/sqlx` + `gorm.io/gorm`）
- **缓存/限流**：Redis（`go-redis/v8`），内置基于 Redis 的分布式限流工具
- **调度**：`robfig/cron/v3`（每日 23:30 建实例）+ 自研延迟队列（`delay_queue` 表）
- **安全**：`golang.org/x/crypto`（bcrypt 密码、SSH 远程执行）
- **对象存储**：`minio-go/v6`（预签名上传）
- **其他**：`etcd`、`mongodb`、`langchaingo`、`otel/jaeger`（扩展中间件，非启动必需）
- **依赖管理**：Go Modules，**vendor 模式**（`-mod=vendor`）

## 架构设计

### 分层架构

```
请求 → router（路由/中间件装配）
      → tool（JWT 认证、租户中间件、日志、OAuth）
      → controller（HTTP 层：参数绑定、响应封装）
      → service（业务逻辑）
      → internal/store（数据访问层，Store 接口）
      → MySQL（公共库 + 各租户库）
```

- `api/request`、`api/response`：接口入参/出参契约
- `models/`：数据模型（user / tenant / dispatch / task / article / chat_msg）
- `pkg/`：跨模块基础设施（`saas_db` 分库连接池、`tenantctx` 租户上下文、`metrics`、`minio`、`cron`、`trace`）
- `utils/`：通用工具（分布式限流、模板渲染、验证、时间轮、wire 等）
- `internal/`：数据访问层（`store`）、可选中间件接入（`cache`/`etcd`/`nosql`/`ai`）、工作流运行时（`workflow`）

### 多租户：分库隔离（Database-per-Tenant）

```
请求头 X-Tenant-Id: default
        │
        ▼
tool.TenantMiddleware
  ├─ 公共库查询 user_tenant：当前用户是否为该租户成员
  ├─ 解析 tenant 表：tenant_id → database_name（如 default → mytest）
  └─ 写入请求 context：tenantctx.WithTenantDB(库名) + WithUserID(用户ID)
        │
        ▼
service 层统一读取 tenantctx.TenantDB(ctx)，经 pkg/saas_db 路由到租户库
```

- **公共库**：`users`、`tenant`、`user_tenant`、`task`、`channel`、`articles` 等
- **租户库**：`nodes`、`line`、`datasource`、`exec_queue`、`node_instance`、`instance_line`、`delay_queue` 等
- 服务启动时 `EnsureDefaultTenant` 幂等初始化默认租户（`default → mytest`）并把内置 admin 绑定为 owner
- 不需要租户上下文的接口（登录、注册、ODS 公共数据源、数据迁移）不挂租户中间件

### 数据管道调度

```
cron（每日 23:30）→ InstanceCreate（从根节点构建 DAG 实例）
      → 节点实例（node_instance）+ 实例连线（instance_line）
      → 延迟队列（delay_queue）驱动下游触发
      → 节点执行写入 exec_queue（运行记录/结果）
```

- 节点保存/连线校验（防成环）→ `nodes` / `line` 表
- 实例状态机：`NotReady → Waiting → Running → Success / Failure / Abort / TimeOut`
- 执行队列 `exec_queue`：`pending → running → success / failed`，`content` 按节点类型区分（`{"sql":...}` / `{"shell":...}`）

## 目录结构

```
myGinServer/
├── main.go                 # 入口：装配 DB、控制器、路由，监听 :8081
├── config.yaml             # 配置文件（DB/对象存储/Redis/SSH）
├── api/                    # 接口契约
│   ├── request/            #   入参结构（含 migrate、dispatch、datasource）
│   └── response/           #   出参结构
├── config/                 # 配置加载（yaml）
├── controller/             # HTTP 控制器
│   ├── user_controller.go  #   用户/租户/任务/文章/对象存储接口
│   ├── dispatch_controller.go # 节点/连线/画布/测试运行/数据源
│   └── migrate_controller.go  # 数据迁移（run/status/stop）
├── service/                # 业务逻辑
│   ├── userserver/         #   用户、任务
│   ├── tenantserver/       #   租户
│   ├── dispatchserver/     #   调度：节点服务、数据源、Shell 执行、迁移配置解析
│   └── migrate/            #   类 DataX 迁移引擎
├── models/                 # 数据模型
│   ├── user/  tenant/  dispatch/  task/  article/  chat_msg/
├── internal/               # 数据访问与运行时
│   ├── store/              #   DBStore 接口 + dispatch store + 延迟队列
│   ├── workflow/           #   节点运行时（DAG 实例执行）
│   └── cache/ etcd/ nosql/ ai/   # 可选中间件接入
├── pkg/                    # 基础设施
│   ├── saas_db/            #   分库连接池（按租户库名缓存 sqlx.DB）
│   ├── tenantctx/          #   租户上下文（租户库名/用户 ID 透传）
│   └── metrics/ minio/ cron/ trace/
├── router/                 # 路由注册、CORS、Recovery、中间件装配
├── tool/                   # JWT 认证、租户/管理员中间件、日志、OAuth
├── utils/                  # 限流、模板渲染、验证、时间轮等
├── init/                   # 数据库初始化 SQL
│   ├── init.sql            #   公共库表（users/channel/articles/message）
│   ├── tenant.sql          #   租户/用户-租户关联 + admin 种子
│   └── dispatch.sql        #   租户库表（nodes/line/datasource/exec_queue/...）
└── docs/                   # 接口文档（dispatch_api.md、migrate_api.md）
```

## 模块说明

### 1. 认证与用户

- **登录**：`POST /auth/login`（JWT），返回 token，后续请求携带 `Authorization: Bearer <token>`
- **密码**：bcrypt 加密存储，`/api/user/password` 修改、管理员 `/api/user/:id/password` 重置（`{new_password}`）
- **注册**：`POST /register`
- **个人中心**：`GET/PUT /api/user/profile`
- **用户管理**（需 admin）：列表/新建/编辑/禁用，见「API 一览」
- 默认账号：`admin / admin`（用户 ID `00000000-0000-0000-0000-000000000001`，角色 admin）

### 2. 多租户

- `GET /api/tenant/list`：当前用户参与的租户列表（含角色）
- 分库隔离：请求头 `X-Tenant-Id`（如 `default`）→ 租户库（如 `mytest`）
- 新增租户即创建新数据库（`database_name` 由后端生成），业务表使用 `init/dispatch.sql`

### 3. 调度引擎

| 接口 | 说明 |
|---|---|
| `POST /api/dispatch/node` | 保存节点（类型 Virtual/SQL/Collect/Sync/Shell） |
| `POST /api/dispatch/line` | 保存连线（自动防成环校验） |
| `DELETE /api/dispatch/node`、`/line` | 删除节点/连线 |
| `GET /api/dispatch/graph` | 获取画布全部节点与连线 |
| `POST /api/dispatch/test_run` | 提交节点测试运行（SQL 或 Shell） |
| `GET /api/dispatch/query_result` | 轮询测试运行结果 |

### 4. 节点类型与 content 结构

| 类型 | content 结构 | 执行语义 |
|---|---|---|
| `SQL` | `{"sql":"...","param":{}}` | 在 ODS 公共数据源上执行 SQL，返回表头+行数据 |
| `Shell` | `{"shell":"..."}` | 经 SSH 在远程主机执行脚本（`bash -s` 经 stdin），返回输出 |
| `Collect` | `{"source":{"datasource_id","table",...},"target":{"table","create_ddl","is_cover"}}` | 从外部数据源采集到当前租户库（目标连接 = 租户库） |
| `Sync` | `{"source":{"datasource_id","table"},"target":{"datasource_id","table","create_ddl","is_cover"}}` | 任意两端 MySQL 同步（迁移载体） |
| `Virtual` | 任意/留空 | 虚拟节点（DAG 根） |

### 5. 数据源

- `POST /api/datasource/save` / `GET /api/datasource/list`：租户库内数据源管理（`conn_str` 为 JSON，存于租户库 `datasource` 表，列表返回脱敏）
- `GET /api/datasource/tables`、`/columns`：按数据源 ID 读取表/列元数据
- ODS 公共数据源：`GET /api/datasource/ods/tables`、`/columns`（固定数据源 ID，不依赖租户）

### 6. Shell 节点（SSH 远程执行）

- 参考 `processor/processor.go`（bigdata-faas）实现：SSH 连接支持**密码 / 私钥（文件路径或内联 PEM）**双认证，`InsecureIgnoreHostKey`，超时控制（默认 5s），脚本经 `bash -s` 由 stdin 送入解释器执行
- 捕获 stdout/stderr/退出码；非零退出码记为失败并在响应 `output` 中返回
- 连接配置见 `config.yaml` 的 `ssh:` 块；单测使用进程内 SSH 服务器（`service/dispatchserver/shell_test.go`）

### 7. 数据迁移（类 DataX，MySQL → MySQL）

- **节点驱动**：连接信息不经过接口，取自已保存的 Collect/Sync 节点 content（`node_id`）
  - Sync：源/目标均为数据源引用，支持任意两端
  - Collect：源 = 外部数据源，目标 = 当前租户库
- **能力**：流式读取 + 分批事务写入（失败整批回滚）；`insert / replace / upsert` 三种模式；`where` 过滤做增量；`channels>1` 按数字主键区间并发分片（无主键自动回退串行）；目标表缺失自动建表（优先节点 `create_ddl`，否则复制源表 DDL）；运行中可取消（批次边界生效）
- 接口：`POST /api/migrate/run`、`GET /api/migrate/status?run_id=`、`POST /api/migrate/stop?run_id=`
- 幂等：`run_id` 运行中重复提交返回 400，已结束任务可覆盖重跑
- 详见 `docs/migrate_api.md`

### 8. 对象存储

- `PUT /api/object/presigned-upload-url`：MinIO 预签名上传地址

### 9. 辅助能力

- 任务管理（`/api/task/*`）、文章（`/api/article*`、`/api/articles`）、渠道（`/api/channels`）
- 分布式限流工具（`utils/distributed_rate_limiter.go`，基于 Redis）
- Prometheus metrics（`pkg/metrics`，`/metrics`，默认关闭，可在 `main.go` 开启后监听 :18081）
- pprof（`/debug/pprof`）、Swagger（`/swagger/index.html`）

## 快速开始

### 环境依赖

- Go 1.24+（vendor 模式构建）
- MySQL 5.7+ / 8.x
- Redis（可选，限流使用）

### 1. 初始化数据库

```bash
# 创建默认库并执行初始化脚本（公共库）
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS mytest DEFAULT CHARSET utf8mb4"
mysql -uroot -p mytest < init/init.sql
mysql -uroot -p mytest < init/tenant.sql
mysql -uroot -p mytest < init/dispatch.sql
```

> `tenant.sql` 会写入默认租户（`default → mytest`）与内置 admin 用户；服务启动时 `EnsureDefaultTenant` 也会幂等补齐，两者取其一即可。新增租户库执行 `init/dispatch.sql`。

### 2. 配置

按需修改 `config.yaml`（数据库连接、Redis、SSH、对象存储）。

### 3. 运行

```bash
# 方式一：直接运行
go run main.go

# 方式二：构建后运行（vendor 模式）
go build -mod=vendor -o mygin_server .
./mygin_server

# 方式三：Docker
docker build -t myginserver .
docker run -p 8081:8081 myginserver
```

服务启动后监听 `http://localhost:8081`，健康检查 `GET /` 返回 `hello world`。

### 4. 验证登录

```bash
curl -X POST http://127.0.0.1:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
# 返回 token，后续请求携带 Authorization: Bearer <token>
# 租户相关接口额外携带 X-Tenant-Id: default
```

## 配置说明（config.yaml）

| 配置块 | 字段 | 说明 |
|---|---|---|
| `db` | `db_host/db_port/db_user/db_password/db_name` | 全局 MySQL 连接（公共库 + 租户库路由基库） |
| `object` | `endpoint/access_id/access_secret/bucket_name/root_path` | MinIO 对象存储 |
| `cache` | `redis_host/redis_port/password/pool_size/db` | Redis（限流等） |
| `ssh` | `host/user/password/rsa_private_key_path/rsa_private_key/passphrase/shell_type/timeout_seconds` | Shell 节点远程执行连接（密码或私钥二选一） |

## API 一览

| 方法 | 路径 | 认证 | 租户 | 说明 |
|---|---|---|---|---|
| POST | `/auth/login` | - | - | 登录获取 JWT |
| POST | `/register` | - | - | 注册 |
| GET | `/callback` | - | - | OAuth 回调 |
| GET | `/swagger/*any` | - | - | Swagger 文档 |
| GET | `/api/user/profile` | ✓ | - | 个人资料 |
| PUT | `/api/user/profile` | ✓ | - | 修改资料 |
| POST | `/api/user/password` | ✓ | - | 修改密码 |
| GET | `/api/tenant/list` | ✓ | - | 我的租户列表 |
| GET/POST/DELETE | `/api/task/*` | ✓ | - | 任务管理 |
| GET | `/api/channels`、`/api/articles`、`/api/article/*` | ✓ | - | 渠道/文章 |
| GET/POST/PUT/DELETE | `/api/user/*` | ✓(admin) | - | 用户管理 |
| PUT | `/api/object/presigned-upload-url` | ✓ | - | MinIO 预签名 |
| POST/DELETE/GET | `/api/dispatch/*` | ✓ | ✓ | 节点/连线/画布/测试运行 |
| GET/POST | `/api/datasource/*` | ✓ | ✓ | 数据源管理/元数据 |
| GET | `/api/datasource/ods/*` | ✓ | - | ODS 公共数据源元数据 |
| POST | `/api/migrate/run` | ✓ | - | 提交迁移任务（node_id 驱动） |
| GET | `/api/migrate/status` | ✓ | - | 迁移状态 |
| POST | `/api/migrate/stop` | ✓ | - | 取消迁移任务 |

## 数据库表

**公共库**：`users`、`tenant`、`user_tenant`、`task`、`channel`、`articles`、`message`

**租户库**（`init/dispatch.sql`）：

| 表 | 说明 |
|---|---|
| `nodes` | 画布节点（content 为 JSON，type 区分 SQL/Collect/Sync/Shell/Virtual） |
| `line` | 画布连线（Dotted/Solid，成环校验） |
| `datasource` | 数据源（conn_str 为 JSON 连接信息） |
| `exec_queue` | 执行队列（content/response/tenant_db，记录每次测试运行） |
| `node_instance` | 节点实例（调度批次内每个节点的执行记录） |
| `instance_line` | 实例连线（批次内上下游） |
| `delay_queue` | 延迟队列（驱动下游实例触发） |

## 测试

```bash
# 运行全部测试
go test ./... -mod=vendor

# 核心模块
go test ./service/migrate/ ./service/dispatchserver/ ./internal/store/delayqueue/ ./utils/ -v
```

覆盖要点：迁移 SQL 生成与分片（`sql_builder_test.go`）、节点迁移配置解析（`migrate_config_test.go`）、Shell 远程执行（进程内 SSH 服务器，`shell_test.go`）、DAG 实例化（`dag_test.go`）。

> 注：`internal/ai`、`internal/etcd`、`internal/nosql`、`pkg/cron`、`pkg/trace` 等包依赖外部中间件（mongo/etcd），无对应服务时这些包的测试会失败，与业务代码无关。

## 已知边界与规划

- 迁移任务记录为**内存态**，进程重启后历史任务丢失；后续可落库持久化
- `channels` 并发依赖源表**单列数字主键**；复合主键/无主键自动回退串行
- ODS 公共数据源 ID 为后端常量（`OdsDatasourceId`），前端直接引用
- 前端画布配套项目：`~/workspace/myReact/myreact-app`（CRA5 + React 19 + React Flow）
