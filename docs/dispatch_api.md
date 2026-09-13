# 调度模块接口文档（api/dispatch/*）

> 面向前端联调使用，说明每个接口的请求地址、请求参数、响应参数及完整示例。

## 1. 基础约定

### 1.1 服务地址

| 环境 | Base URL |
|------|----------|
| 本地 | `http://localhost:8081` |

### 1.2 鉴权说明

`/api/dispatch/*` 路由组**未挂载 JWT 中间件**，即调用时**无需携带 `Authorization` 请求头**，可直接访问。
（与其他 `/api/*` 业务接口不同，后者需要 JWT 认证。）

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

---

## 3. 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/dispatch/node` | 保存 / 新增节点（UPSERT） |
| POST | `/api/dispatch/line` | 保存 / 新增连线（含成环校验） |
| GET | `/api/dispatch/graph` | 获取整个图（全部节点 + 连线） |

---

## 4. POST /api/dispatch/node —— 保存节点

根据 `id` 判断新增或更新：`id` 有值且存在则更新该节点，否则新增。
保存时自动补充：`status`（节点状态）、`deleted=0`、`created_on`、`created_by=admin`。

### 4.1 请求参数（Body，JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 否 | 节点主键 ID（唯一）。有值则更新，为空则新增 |
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
```golang 
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

---

## 5. POST /api/dispatch/line —— 保存连线

根据 `id` 判断新增或更新。保存前会进行**成环校验**：若 `ahead_id == behind_id`，或新增该连线后图中存在回路，则拒绝保存。

### 5.1 请求参数（Body，JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 否 | 连线 ID（唯一）。有值则更新，为空则新增 |
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

## 7. 前端联调要点

1. **鉴权**：`/api/dispatch/*` 无需 JWT，但请求需通过 CORS（已全局开启，允许 `POST/GET/OPTIONS` 等）。
2. **字段命名**：请求与响应均使用 **snake_case**（如 `ahead_id`、`behind_id`、`created_on`）。
3. **content 是字符串**：前端保存节点时，需将内容对象 `JSON.stringify` 后再放入 `content`；读取图数据后需 `JSON.parse` 再使用。
4. **UPSERT 语义**：`id` 传值即更新，不传（空字符串）即新增。建议前端为每个节点/连线生成并维护稳定 ID。
5. **成环校验**：保存连线前前端可先本地校验 `ahead_id != behind_id`，服务端会兜底校验完整环路。
6. **Cron 校验**：节点 `schedule` 非法会返回 500，建议前端先做格式校验或捕获该错误并提示用户。
7. **必填与枚举校验**：节点 `code`/`name`/`type`/`schedule`、连线 `ahead_id`/`behind_id`/`type` 为必填；`type` 取值必须在枚举内（节点 `Virtual/SQL/Collect/Sync`，连线 `Dotted/Solid`），否则接口返回 400。
