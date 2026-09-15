# 调度模块接口文档（api/dispatch/* 与 api/datasource/*）

> 面向前端联调使用，说明每个接口的请求地址、请求参数、响应参数及完整示例。

## 1. 基础约定

### 1.1 服务地址

| 环境 | Base URL |
|------|----------|
| 本地 | `http://localhost:8081` |

### 1.2 鉴权说明

`/api/dispatch/*` 与 `/api/datasource/*` 路由组**均未挂载 JWT 中间件**，即调用时**无需携带 `Authorization` 请求头**，可直接访问。
（与其他 `/api/*` 业务接口不同，后者需要 JWT 认证。）

> ⚠️ **安全提示**：由于无鉴权，且 `GET /api/datasource/list` 会原样返回 `conn_str`（内含数据库账号与明文密码），
> 当前这些接口在公网可被任意调用。上线前需为该路由组补挂鉴权中间件，并对 `conn_str` 做脱敏。

### 1.3 请求与响应编码

- 请求体：`Content-Type: application/json`
- 响应体：`application/json; charset=utf-8`
- 编码：UTF-8

### 1.4 统一响应封装

所有接口均返回统一结构：

**成功响应**（HTTP 200）：

```json
{
  "code": 200,
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | number | 状态码，200 表示成功 |
| data | object / string | 业务数据，见各接口定义 |

**失败响应**（HTTP 状态码对应 `code` 字段）：

```json
{
  "code": 500,
  "message": "错误信息"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | number | 状态码，400 / 500 等 |
| message | string | 错误描述 |

---

## 2. 枚举值说明

### 2.1 节点类型（type）

| 值 | 说明 |
|----|------|
| `Virtual` | 虚拟节点 |
| `SQL` | SQL 节点 |
| `Collect` | 采集节点 |
| `Sync` | 同步节点 |

### 2.2 连线类型（type）

| 值 | 说明 |
|----|------|
| `Dotted` | 点虚线 |
| `Solid` | 实线 |

### 2.3 Cron 表达式（schedule）

使用标准 **5 字段** Cron 表达式，格式：`分 时 日 月 周`。示例：

| 表达式 | 含义 |
|--------|------|
| `*/5 * * * *` | 每 5 分钟执行一次 |
| `0 0 * * *` | 每天 0 点执行一次 |
| `0 9 * * 1-5` | 工作日（周一~周五）9 点执行 |

> 保存节点时会校验 `schedule`，非法表达式将返回错误。

### 2.4 数据源类型（type）

| 值 | 说明 |
|----|------|
| `mysql` | MySQL 数据源（默认） |

> `conn_str` 连接信息结构随 `type` 不同，MySQL 类型的结构见 7.1.2。

---

## 3. 接口列表

### 3.1 调度接口（/api/dispatch/*）

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| POST | `/api/dispatch/node` | 保存 / 新增节点（UPSERT） | §4 |
| DELETE | `/api/dispatch/node` | 逻辑删除节点 | §4.5 |
| POST | `/api/dispatch/line` | 保存 / 新增连线（含成环校验） | §5 |
| DELETE | `/api/dispatch/line` | 逻辑删除连线 | §5.4 |
| GET | `/api/dispatch/graph` | 获取整个图（全部节点 + 连线） | §6 |
| POST | `/api/dispatch/test_run` | 节点测试运行（提交 SQL 异步执行） | §7 |
| GET | `/api/dispatch/query_result` | 查询测试运行结果（**当前未挂载路由**） | §7.4 |

### 3.2 数据源接口（/api/datasource/*）

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| POST | `/api/datasource/save` | 保存 / 新增数据源（UPSERT） | §8.1 |
| GET | `/api/datasource/list` | 获取数据源列表 | §8.2 |
| GET | `/api/datasource/tables` | 按数据源 ID 获取表列表 | §8.3 |
| GET | `/api/datasource/columns` | 按数据源 ID + 表名获取列列表 | §8.4 |
| GET | `/api/datasource/ods/tables` | 获取 ODS 表列表（数据源 ID 由后端固定） | §8.5 |
| GET | `/api/datasource/ods/columns` | 按表名获取 ODS 列列表 | §8.6 |

> `DELETE` 方法用于删除接口，这是本模块唯一使用 DELETE 的场景（路由 `dispatchApi.DELETE("/node")` / `dispatchApi.DELETE("/line")`）。

---

## 4. POST /api/dispatch/node —— 保存节点

根据 `id` 判断新增或更新：`id` 有值且存在则更新该节点，否则新增。
保存时自动补充：`status`（节点状态）、`deleted=0`、`created_on`、`created_by=admin`。

### 4.1 请求参数（Body，JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 节点主键 ID（唯一），保存逻辑为**按主键 upsert**。新增时也**必须**由前端生成 UUID 传入：后端没有「为空则生成」的兜底，传空串会以空主键 INSERT，第二次提交即主键冲突 |
| code | string | 是 | 节点编码 |
| name | string | 是 | 节点名称 |
| content | string | 否 | 节点内容，为 JSON 字符串（结构见 4.3） |
| type | string | 是 | 节点类型，取值见 2.1 |
| schedule | string | 是 | Cron 表达式（5 字段），见 2.3 |

### 4.2 请求示例

```json
{
  "id": "node-001",
  "code": "collect_order",
  "name": "订单采集",
  "content": "{\"source\":{\"datasource_id\":\"ds-1\",\"table\":\"orders\",\"database\":\"biz\"},\"target\":{\"table\":\"dwd_orders\"}}",
  "type": "Collect",
  "schedule": "*/5 * * * *"
}
```
``` golang 
// sql 节点内容结构
type SqlNodeContent struct {
	Sql   string         `json:"sql" db:"sql"`
	Param map[string]any `json:"param" db:"param"`
}

type Column struct {
	Name      string `json:"name"`
	DataType  string `json:"data_type"`
	IsPrimary int8   `json:"is_primary"`
}

type CollectSource struct {
	DatasourceId string   `json:"datasource_id" db:"datasource_id"`
	Table        string   `json:"table" db:"table"`
	Database     string   `json:"database" db:"database"`
	Columns      []Column `json:"columns" db:"columns"`
	CustomSQL    string   `json:"custom_sql" db:"custom_sql"`
}
type CollectTarget struct {
	Table     string   `json:"table"`
	CreateDDL string   `json:"create_ddl"`
	Columns   []Column `json:"columns" db:"columns"`
	IsCover   int8     `json:"is_cover" db:"is_cover"`
}
// 采集节点content 结构
type CollectNodeContent struct {
	Source CollectSource `json:"source"`
	Target CollectTarget `json:"target"`
}

```

### 4.3 content 字段结构说明

`content` 是**字符串类型的 JSON**，按节点 `type` 不同，内容结构不同：

**type = SQL**（`SqlNodeContent`）：

```json
{
  "sql": "SELECT * FROM orders WHERE id = :id",
  "param": { "id": 1001 }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| sql | string | SQL 语句 |
| param | object | SQL 参数（键值对） |

**type = Collect**（`CollectNodeContent`）：

```json
{
  "source": {
    "datasource_id": "ds-1",
    "table": "orders",
    "database": "biz",
    "columns": [
      { "name": "id", "data_type": "bigint", "is_primary": 1 },
      { "name": "name", "data_type": "varchar", "is_primary": 0 }
    ],
    "custom_sql": ""
  },
  "target": {
    "table": "dwd_orders",
    "create_ddl": "",
    "columns": [
      { "name": "id", "data_type": "bigint", "is_primary": 1 }
    ],
    "is_cover": 1
  }
}
```

`source`（采集来源）字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| datasource_id | string | 数据源 ID |
| table | string | 源表名 |
| database | string | 源库名 |
| columns | array | 字段列表，见下表 |
| custom_sql | string | 自定义 SQL |

`target`（采集目标）字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| table | string | 目标表名 |
| create_ddl | string | 建表 DDL |
| columns | array | 字段列表 |
| is_cover | number | 是否覆盖，0=否 1=是 |

`columns` 元素字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 字段名 |
| data_type | string | 字段类型 |
| is_primary | number | 是否主键，0=否 1=是 |

> `type = Virtual` / `Sync` 时，`content` 无固定结构，可传任意 JSON 字符串或留空。

### 4.4 响应示例

**成功**：

```json
{
  "code": 200,
  "data": "ok"
}
```

**失败（参数解析错误）**：

```json
{
  "code": 400,
  "message": "无效的请求数据: ..."
}
```

**失败（Cron 表达式非法 / 服务异常）**：

```json
{
  "code": 500,
  "message": "未知的cron表达式: ..."
}
```

### 4.5 DELETE /api/dispatch/node —— 删除节点

逻辑删除（软删除）：将节点的 `deleted` 字段置为 `1`，数据仍保留在库中，`/api/dispatch/graph` 不再返回。

#### 4.5.1 请求参数（Query）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 要删除的节点 ID |

> 注意：`id` 未做 `required` 校验。若缺失或传空字符串，接口仍返回成功（`data: "ok"`），但实际不会更新任何行。前端需自行保证传值。

#### 4.5.2 请求示例

```
DELETE /api/dispatch/node?id=node-001
```

#### 4.5.3 响应示例

**成功**：

```json
{
  "code": 200,
  "data": "ok"
}
```

**失败（服务异常）**：

```json
{
  "code": 500,
  "message": "错误信息"
}
```

---

## 5. POST /api/dispatch/line —— 保存连线

根据 `id` 判断新增或更新。保存前会进行**成环校验**：若 `ahead_id == behind_id`，或新增该连线后图中存在回路，则拒绝保存。

### 5.1 请求参数（Body，JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 连线 ID（唯一），保存逻辑为按主键 upsert。同 §4.1，新增时也必须传入前端生成的 UUID |
| ahead_id | string | 是 | 前驱节点 ID |
| behind_id | string | 是 | 后继节点 ID |
| type | string | 是 | 连线类型，取值见 2.2 |

### 5.2 请求示例

```json
{
  "id": "line-001",
  "ahead_id": "node-001",
  "behind_id": "node-002",
  "type": "Solid"
}
```

### 5.3 响应示例

**成功**：

```json
{
  "code": 200,
  "data": "ok"
}
```

**失败（参数解析错误）**：

```json
{
  "code": 400,
  "message": "无效的请求参数"
}
```

**失败（成环 / 服务异常）**：

```json
{
  "code": 500,
  "message": "成环异常: node-002->node-001"
}
```

### 5.4 DELETE /api/dispatch/line —— 删除连线

逻辑删除（软删除）：将连线的 `deleted` 字段置为 `1`。删除后该连线不再参与 `/api/dispatch/graph` 渲染与成环校验。

#### 5.4.1 请求参数（Query）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 要删除的连线 ID |

> 与删除节点相同，`id` 未做 `required` 校验，传空会「成功但无实际影响」。

#### 5.4.2 请求示例

```
DELETE /api/dispatch/line?id=line-001
```

#### 5.4.3 响应示例

**成功**：

```json
{
  "code": 200,
  "data": "ok"
}
```

**失败（服务异常）**：

```json
{
  "code": 500,
  "message": "错误信息"
}
```

---

## 6. GET /api/dispatch/graph —— 获取整个图

返回当前全部未删除的节点与连线，用于前端一次性渲染调度流程图。

### 6.1 请求参数

无（无需请求体，无需 Query 参数）。

### 6.2 响应参数（data 字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| nodes | array | 节点列表，元素结构见下表 |
| lines | array | 连线列表，元素结构见下表 |

`nodes` 元素：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 节点 ID |
| code | string | 节点编码 |
| name | string | 节点名称 |
| content | string | 节点内容（JSON 字符串） |
| type | string | 节点类型 |
| schedule | string | Cron 表达式 |

`lines` 元素：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 连线 ID |
| ahead_id | string | 前驱节点 ID |
| behind_id | string | 后继节点 ID |
| type | string | 连线类型 |

### 6.3 响应示例

```json
{
  "code": 200,
  "data": {
    "nodes": [
      {
        "id": "node-001",
        "code": "collect_order",
        "name": "订单采集",
        "content": "{\"source\":{...},\"target\":{...}}",
        "type": "Collect",
        "schedule": "*/5 * * * *"
      },
      {
        "id": "node-002",
        "code": "sync_user",
        "name": "用户同步",
        "content": "",
        "type": "Sync",
        "schedule": "0 0 * * *"
      }
    ],
    "lines": [
      {
        "id": "line-001",
        "ahead_id": "node-001",
        "behind_id": "node-002",
        "type": "Solid"
      }
    ]
  }
}
```

### 6.4 失败响应

```json
{
  "code": 500,
  "message": "错误信息"
}
```

---

## 7. 节点测试运行接口（/api/dispatch/test_run）

用于在调度流程之外**试跑一段 SQL**，校验 SQL 与参数是否正确。

该接口为**异步**接口：提交后立即返回 `ok`，SQL 的实际执行结果需要另行查询（见 7.6）。

> ✅ **当前版本该接口可用**：`RunId` 上曾有的 `binding:"run_id"`（会触发 validator 内部 `panic` 拖垮进程）已改为
> `binding:"required"`；执行链路也曾把 `entity.RunId` 当作 SQL 执行，现已改为 `entity.Sql`。
> 相关历史缺陷（D1、D3、D15）均已修复，可正常联调。

### 7.1 请求参数（Body，JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| run_id | string | 是 | 本次运行的唯一 ID，**幂等键**。相同 `run_id` 重复提交只会执行一次 |
| sql | string | 是 | **Base64 编码**的 SQL 模板（不是明文 SQL） |
| params | object | 是 | SQL 模板参数，用于替换 SQL 中的 `{{key}}` 占位符。**不可缺省**（缺省返回 400）；但传空对象 `{}` 可以通过校验，此时模板中若有占位符会在渲染阶段报「参数缺失」 |
| type | string | 否 | 预留字段，当前不参与逻辑 |

> `run_id` 建议使用 UUID。该字段对应 `exec_queue.run_id`（`VARCHAR(32)` 且带唯一索引），超长或重复都会写库失败。

### 7.2 SQL 模板与参数规则

`sql` 字段需先做 Base64 编码，服务端解码后再按 `{{key}}` 占位符替换。值会按类型转成 SQL 字面量：

| 参数类型 | 渲染结果 |
|----------|----------|
| string | `'abc'`（自动加单引号并转义） |
| number / bool | 直接输出（bool 转 1/0） |
| null | `NULL` |
| 表名/列名（需后端标记为 Identifier） | `` `tbl` ``（反引号包裹） |

占位符缺少对应参数时会直接报错，不会静默生成错误 SQL。

**明文 SQL 示例（编码前）**：

```sql
SELECT * FROM orders WHERE dt = '{{dt}}' AND id = {{id}} LIMIT 10
```

对应 `params`：

```json
{ "dt": "2026-09-15", "id": 1001 }
```

### 7.3 请求示例

```json
{
  "run_id": "a3f1c2d4-0001-4f2b-9c31-8e7d6a5b4c3d",
  "sql": "U0VMRUNUICogRlJPTSBvcmRlcnMgV0hFUkUgZHQgPSAne3tkdH19JyBMSU1JVCAxMA==",
  "params": { "dt": "2026-09-15" },
  "type": "SQL"
}
```

### 7.4 响应示例

**成功（已受理，非执行完成）**：

```json
{
  "code": 200,
  "data": "ok"
}
```

**失败（参数校验不通过）**：

```json
{
  "code": 400,
  "message": "..."
}
```

**失败（SQL 解码 / 参数替换 / 入队异常）**：

```json
{
  "code": 500,
  "message": "测试运行失败: SQL解码失败: ..."
}
```

**失败（执行队列已满，容量 100）**：

```json
{
  "code": 500,
  "message": "测试运行失败: 测试运行队列已满（容量 100），请稍后重试"
}
```

> 队列满时**立即返回该错误**，不会阻塞等待槽位释放。此时第 4 步已写入的那条
> `exec_queue` 记录会被就地置为 `failed` 并写明原因，前端轮询 `query_result`
> 能正常拿到终止状态，不会一直空转（见 §7.5）。

### 7.5 异步执行流程

```
POST /api/dispatch/test_run
   │  1. run_id 幂等校验（已存在则直接返回 ok）
   │  2. Base64 解码 sql
   │  3. {{key}} 参数替换
   │  4. 写入 exec_queue（status = pending）
   │  5. 投递到内存 channel
   │     └─ 投递失败（队列已满 / ctx 已取消）时：把第 4 步那条记录直接置为 failed
   │        并写入失败原因，接口返回 500；该记录不会停留在 pending
   ▼
立即返回 {"code":200,"data":"ok"}
   │
   │  （后台 worker 消费）
   │  6. exec_queue.status → running
   │  7. 连接 ODS 数据源（后端固定数据源）执行 SQL
   │  8. 回写 exec_queue.response 与 status → success / failed
   ▼
GET /api/dispatch/query_result?run_id=...   ← 取结果
```

`exec_queue.status` 取值：

| 值 | 说明 |
|----|------|
| `pending` | 待执行（已入队） |
| `loading` | 预留 |
| `running` | 执行中 |
| `success` | 执行成功 |
| `failed` | 执行失败 |

### 7.6 GET /api/dispatch/query_result —— 查询测试运行结果

> ✅ **该路由已挂载**（`router/router.go:130`：`dispatchApi.GET("/query_result", dispatch.QueryResult)`），可正常访问。

#### 7.6.1 请求参数（Query）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| run_id | string | 是 | 提交测试运行时使用的 `run_id` |

#### 7.6.2 响应参数（data 字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| msg | string | 执行信息，成功为 `success`，失败为错误描述 |
| sql | string | 实际执行的 SQL 原文（已由 Base64 解码并完成 `{{key}}` 参数替换） |
| header | array | 结果集列名列表 |
| body | array | 结果集数据，元素为「列名 → 值」的对象 |

> 当 `exec_queue.status` 仍为 `pending` / `running`（即未结束）时，`data` 返回**空对象** `{}`，前端应据此判断「仍在执行」并轮询。

#### 7.6.3 响应示例

**执行完成**：

```json
{
  "code": 200,
  "data": {
    "msg": "success",
    "sql": "SELECT * FROM orders WHERE dt = '2026-09-15' LIMIT 10",
    "header": ["id", "name", "amount"],
    "body": [
      { "id": 1001, "name": "张三", "amount": "99.00" }
    ]
  }
}
```

**执行失败**：

```json
{
  "code": 200,
  "data": {
    "msg": "Error 1146: Table 'biz.orders' doesn't exist",
    "sql": "SELECT * FROM orders LIMIT 10",
    "header": null,
    "body": null
  }
}
```

**仍在执行中**：

```json
{
  "code": 200,
  "data": {}
}
```

#### 7.6.4 失败响应

```json
{
  "code": 500,
  "message": "运行结果获取失败: ..."
}
```

---

## 8. 数据源接口（api/datasource/*）

数据源用于为采集节点（`Collect`）提供连接信息，路由同样未挂 JWT。

### 8.1 POST /api/datasource/save —— 保存数据源

根据 `id` 判断新增或更新。保存时自动补充：`created_on`、`created_by=admin`。

#### 8.1.1 请求参数（Body，JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 数据源 ID（唯一），保存逻辑为按主键 upsert。同 §4.1，新增时也必须传入前端生成的 UUID |
| code | string | 是 | 数据源编码 |
| name | string | 是 | 数据源名称 |
| type | string | 是 | 数据源类型，见 2.4（如 `mysql`） |
| conn_str | string | 是 | 连接信息，JSON 字符串（结构见 7.1.2） |

#### 8.1.2 conn_str 字段结构说明

`conn_str` 是**字符串类型的 JSON**。`type = mysql` 时（`MysqlDatasourceContent`）：

```json
{
  "user": "root",
  "passwd": "123456",
  "host": "127.0.0.1",
  "port": 3306,
  "database": "biz",
  "schema": "public"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| user | string | 用户名 |
| passwd | string | 密码 |
| host | string | 主机地址 |
| port | number | 端口号 |
| database | string | 数据库名 |
| schema | string | Schema 名 |

#### 8.1.3 请求示例

```json
{
  "id": "ds-1",
  "code": "mysql_biz",
  "name": "业务库",
  "type": "mysql",
  "conn_str": "{\"user\":\"root\",\"passwd\":\"123456\",\"host\":\"127.0.0.1\",\"port\":3306,\"database\":\"biz\",\"schema\":\"public\"}"
}
```

#### 8.1.4 响应示例

**成功**：

```json
{
  "code": 200,
  "data": "ok"
}
```

**失败（参数解析错误）**：

```json
{
  "code": 400,
  "message": "获取请求参数失败"
}
```

**失败（保存异常）**：

```json
{
  "code": 500,
  "message": "保存数据源失败: ..."
}
```

### 8.2 GET /api/datasource/list —— 获取数据源列表

返回全部数据源。

#### 8.2.1 请求参数

无。

#### 8.2.2 响应参数（data 字段）

`data` 为数组，元素结构如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 数据源 ID |
| code | string | 数据源编码 |
| name | string | 数据源名称 |
| type | string | 数据源类型 |
| conn_str | string | 连接信息（JSON 字符串） |

#### 8.2.3 响应示例

```json
{
  "code": 200,
  "data": [
    {
      "id": "ds-1",
      "code": "mysql_biz",
      "name": "业务库",
      "type": "mysql",
      "conn_str": "{\"user\":\"root\",\"passwd\":\"123456\",\"host\":\"127.0.0.1\",\"port\":3306,\"database\":\"biz\",\"schema\":\"public\"}"
    }
  ]
}
```

#### 8.2.4 失败响应

```json
{
  "code": 500,
  "message": "获取数据源列表失败: ..."
}
```

### 8.3 GET /api/datasource/tables —— 按数据源 ID 获取表列表

根据数据源 ID 连接其指向的 MySQL 数据库，返回该库下所有表的表名与注释。

#### 8.3.1 请求参数（Query）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ds_id | string | 是 | 数据源 ID（唯一） |

#### 8.3.2 响应参数（data 字段）

`data` 为数组，元素结构如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 表名 |
| description | string | 表注释（无注释时为空字符串） |

#### 8.3.3 响应示例

```json
{
  "code": 200,
  "data": [
    { "name": "orders", "description": "订单表" },
    { "name": "users", "description": "用户表" }
  ]
}
```

#### 8.3.4 失败响应

```json
{
  "code": 400,
  "message": "缺少数据源ID参数"
}
```

```json
{
  "code": 500,
  "message": "获取数据源失败: ... / 查询表列表失败: ..."
}
```

### 8.4 GET /api/datasource/columns —— 按数据源 ID + 表名获取列列表

根据数据源 ID 与表名，返回该表下所有列的字段名、类型、注释及主键序号。

#### 8.4.1 请求参数（Query）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ds_id | string | 是 | 数据源 ID（唯一） |
| table | string | 是 | 表名 |

#### 8.4.2 响应参数（data 字段）

`data` 为数组，元素结构如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 字段名 |
| type | string | 字段类型（如 `varchar(64)`、`bigint`） |
| comment | string | 字段注释（无注释时为空字符串） |
| value | string | 可选，分区字段 value 值（当前固定为空） |
| primary_key_seq | number | 主键序号，0 表示非主键，1 起为复合主键中的第几个字段 |

#### 8.4.3 响应示例

```json
{
  "code": 200,
  "data": [
    { "name": "id", "type": "bigint", "comment": "主键", "value": "", "primary_key_seq": 1 },
    { "name": "name", "type": "varchar(64)", "comment": "姓名", "value": "", "primary_key_seq": 0 }
  ]
}
```

#### 8.4.4 失败响应

```json
{
  "code": 400,
  "message": "缺少数据源ID或表名参数"
}
```

```json
{
  "code": 500,
  "message": "查询列信息失败: ..."
}
```

### 8.5 GET /api/datasource/ods/tables —— 获取 ODS 表列表

数据源 ID 由**后端固定**（常量 `OdsDatasourceId`），前端无需传入。

#### 8.5.1 请求参数（Query）

无。

#### 8.5.2 响应参数（data 字段）

`data` 为数组，元素结构如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 表名 |
| description | string | 表注释（无注释时为空字符串） |

#### 8.5.3 响应示例

```json
{
  "code": 200,
  "data": [
    { "name": "orders", "description": "订单表" },
    { "name": "users", "description": "用户表" }
  ]
}
```

#### 8.5.4 失败响应

```json
{
  "code": 500,
  "message": "获取数据源失败: ... / 查询表列表失败: ..."
}
```

### 8.6 GET /api/datasource/ods/columns —— 按 ODS 表名获取列列表

数据源 ID 由**后端固定**（常量 `OdsDatasourceId`）。

#### 8.6.1 请求参数（Query）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| table | string | 是 | 表名 |

#### 8.6.2 响应参数（data 字段）

`data` 为数组，元素结构如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 字段名 |
| type | string | 字段类型（如 `varchar(64)`、`bigint`） |
| comment | string | 字段注释（无注释时为空字符串） |
| value | string | 可选，分区字段 value 值（当前固定为空） |
| primary_key_seq | number | 主键序号，0 表示非主键，1 起为复合主键中的第几个字段 |

#### 8.6.3 响应示例

```json
{
  "code": 200,
  "data": [
    { "name": "id", "type": "bigint", "comment": "主键", "value": "", "primary_key_seq": 1 },
    { "name": "name", "type": "varchar(64)", "comment": "姓名", "value": "", "primary_key_seq": 0 }
  ]
}
```

#### 8.6.4 失败响应

```json
{
  "code": 400,
  "message": "缺少数据源ID或表名参数"
}
```

```json
{
  "code": 500,
  "message": "查询列信息失败: ..."
}
```
---

## 9. 前端联调要点

1. **鉴权**：`/api/dispatch/*` 无需 JWT，但请求需通过 CORS（已全局开启，允许 `POST/GET/OPTIONS/PUT/PATCH/DELETE`）。
2. **字段命名**：请求与响应均使用 **snake_case**（如 `ahead_id`、`behind_id`、`created_on`）。
3. **content 是字符串**：前端保存节点时，需将内容对象 `JSON.stringify` 后再放入 `content`；读取图数据后需 `JSON.parse` 再使用。
4. **UPSERT 语义**：`id` 传值即更新，不传（空字符串）即新增。建议前端为每个节点/连线生成并维护稳定 ID。
5. **成环校验**：保存连线前前端可先本地校验 `ahead_id != behind_id`，服务端会兜底校验完整环路。
6. **Cron 校验**：节点 `schedule` 非法会返回 500，建议前端先做格式校验或捕获该错误并提示用户。
7. **必填与枚举校验**：节点 `code`/`name`/`type`/`schedule`、连线 `ahead_id`/`behind_id`/`type` 为必填；`type` 取值必须在枚举内（节点 `Virtual/SQL/Collect/Sync`，连线 `Dotted/Solid`），否则接口返回 400。
8. **数据源校验与 conn_str**：数据源 `code`/`name`/`type`/`conn_str` 为必填；`conn_str` 同样是 JSON 字符串，保存时 `JSON.stringify`、读取时 `JSON.parse`。
9. **删除接口**：`DELETE /api/dispatch/node` 与 `DELETE /api/dispatch/line` 使用 **Query 参数** `id`（不是 Body），且为逻辑删除。删除时建议前端先做二次确认。
10. **测试运行是异步的**：`POST /api/dispatch/test_run` 返回 `ok` 仅代表「已受理」，不代表 SQL 执行成功。`sql` 必须 Base64 编码，`params` 不可缺省（`{}` 可通过）。执行结果需通过 `GET /api/dispatch/query_result?run_id=...` 轮询获取（**该路由当前尚未挂载，见 7.6**）。

---

## 10. 已知缺陷（联调前必读）

下表是当前代码版本的接口可用性清单，兼作本轮审计的回归记录。标 ✅ 的已修复并复核通过，
标 ❌ 的**仍然存在，请勿按「正常路径」预期联调**。

当前唯一未闭合项为 **D16**：`SendEntity` 的阻塞投递子项已作为 **D17** 修复，
剩下的只有「进程重启后未完成的任务无法恢复」，仅影响重启后的历史任务。

| 编号 | 状态 | 影响接口 | 现象 | 根因位置 |
|------|------|----------|------|----------|
| D1 | ✅ 已修复 | `POST /api/dispatch/test_run` | ~~请求必崩进程~~ | `TestRunSqlReq.RunId` 的 `binding` 已由 `run_id` 改为 `required` |
| D2 | ✅ 已修复 | `GET /api/dispatch/query_result` | ~~恒定 404~~ | `router/router.go:130` 已注册 `dispatchApi.GET("/query_result", dispatch.QueryResult)` |
| D3 | ✅ 已修复 | `POST /api/dispatch/test_run` 的实际执行 | ~~SQL 永不执行，结果恒为 `failed`~~ | `service/dispatchserver/dispatch.go:120` 已改为 `entity.Sql`；`node_server.go:152` 的 `SendEntity(ctx, req.RunId, template)` 传参顺序也正确。残留的响应字段回显问题另记为 D15 |
| D4 | ✅ 已修复 | 全部依赖数据源的接口 | ~~数据源报错时进程退出~~ | 两处 `log.Fatalf` 已移除，改为直接 `return` |
| D5 | ✅ 已修复 | 延时队列全部投递 | ~~插入 `delay_queue` 必然失败~~ | `init/dispatch.sql` 的 `delay_queue` 建表语句已补上 `topic VARCHAR(64) NOT NULL` 与 `max_retry INT NOT NULL DEFAULT 3` |
| D6 | ✅ 已修复 | 首次初始化数据库 | ~~建表脚本执行报语法错误~~ | `init/dispatch.sql`：`exec_queue.status` 行末已补逗号；`insert into datasource ...` 语句末已补分号 |
| D7 | ✅ 已修复 | `POST /api/dispatch/node`、`POST /api/dispatch/line` | ~~未传 `id` 时新增第二次会主键冲突~~ | `NodeSaveReq.Id` / `LineSaveReq.Id` / `DatasourceReq.Id` 均已加 `binding:"required"`；契约变为「新增也必须传前端生成的 UUID」，详见 §4.1/§5.1/§8.1 |
| D8 | ✅ 已修复 | `POST /api/datasource/save` | ~~首次（`id` 为空）保存失败~~ | 同 D7，`DatasourceReq.Id` 已必填 |
| D9 | ✅ 已修复 | DAG 实例化（`InstanceCreate`） | ~~写入 `node_instance` 可能失败~~ | `init/dispatch.sql` 的 `start_time`/`end_time` 已改为 `NULL DEFAULT NULL`（实例创建时尚未开始执行，本就不该是 `NOT NULL`），`models/dispatch/dispatch.go` 的 gorm tag 同步去掉 `NOT NULL` |
| D10 | ✅ 已修复 | DAG 实例化（`InstanceCreate`） | ~~根节点在批次窗口内生成 0 个实例，进而 panic~~ | `utils/cron_parse.go` 的取值循环直接用 `start` 起步，而 cron 的 `Next` 返回**严格大于**入参的时间点，导致落在窗口左端点上的触发（如 `0 0 * * *`）被整条吞掉；已改为起步前回退 1 秒 |
| D11 | ✅ 已修复 | 下游实例触发 | ~~延时值算错，下游被立即执行~~ | `Second()` 是分钟内的秒序号，`v.Second() - time.Now().Second()` 得不到时间差；已统一改为 `time.Until(executeTime)` |
| D12 | ✅ 已修复 | 下游实例触发 | ~~对 nil `producer` 调用 `Publish`~~ | `controller/dispatch_controller.go:43` 已改为 `NewNodeRuntime(iStore, producer, consumer)`，`node_runtime.go:91` 的调用点不再为空 |
| D13 | ✅ 已修复 | 下游实例触发 | ~~处理器内 panic 会终止**整个进程**~~ | `internal/store/delayqueue/consumer.go` 新增 `callHandler`，用 `defer recover()` 把 panic 收敛为单次调用的 error，`processOne` 改走它。panic 现在只会让当前这条消息按失败重试，不再打挂进程 |
| D14 | ✅ 已修复 | 下游实例触发 | ~~启动瞬间的待处理消息被误判 `failed`；并有数据竞争~~ | `controller/dispatch_controller.go` 已把 `StartWorkers(ctx, 4)` 挪到 `Dispatch(...)`（内部 `Register`）**之后**；`handlers` 加 `sync.RWMutex`，`Register` 走 `Lock`、读取走新增的 `lookupHandler`（`RLock`）。竞争可由 `go test -race ./internal/store/delayqueue/` 复核 |
| D15 | ✅ 已修复 | `GET /api/dispatch/query_result` | ~~`sql` 字段回显的是 `run_id`~~ | `service/dispatchserver/dispatch.go` 已改为 `resp.Sql = entity.Sql` |
| D16 | ❌ | `POST /api/dispatch/test_run`（仅影响进程重启后的历史任务） | 重启后未完成的测试运行**永久卡在 `pending`** | `service/dispatchserver/dispatch.go` 的 `init()` 只有一句 `// todo 读取数据库中的未结束的exec_queue中的数据写入到channel`，尚未实现——进程重启后内存 channel 里未消费的任务全部丢失，对应 `exec_queue` 行的 `status` 永远是 `pending`，前端会一直轮询到空对象 `{}`。原先并列的「提交量突增时请求阻塞」子项已拆出并作为 D17 修复 |
| D17 | ✅ 已修复 | `POST /api/dispatch/test_run` | ~~并发超 100 时请求阻塞在 handler 内且无法随客户端断开解除；投递失败的记录永远 `pending`~~ | ① `SendEntity` 原 `select` 的 `default` 分支里是**阻塞**发送 `ch <- ...`，且完全没监听 `ctx`，队列满即挂死；已改为「发送 / `ctx.Done()` 竞争，`default` 立即返回『队列已满』错误」，并在入口补一次 `ctx.Err()` 前置检查——否则 ctx 已取消而 channel 恰有空位时，`select` 会在两个就绪分支间随机选择，约 50% 概率仍投递已取消的请求。② `node_server.go` 的 `TestRun` 在 `SendEntity` 失败时新增补偿写入，把该 `exec_queue` 行置为 `failed` 并写明原因，让前端轮询能终止；补偿写入必须用 `context.WithoutCancel(ctx)` 派生，因为投递失败的原因常常正是 ctx 被取消。回归测试见 `service/dispatchserver/dispatch_test.go` |

固定的枚举值、字段长度限制请以本文档第 2 节与各接口章节为准。
