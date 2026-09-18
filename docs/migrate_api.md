# 数据迁移（类 DataX）接口说明

`POST /api/migrate/run`、`GET /api/migrate/status`、`POST /api/migrate/stop`

> 本模块实现 MySQL → MySQL 表级数据迁移（流式读取 + 分批事务写入），
> 能力对齐 DataX 的常用场景：全量 / 增量（where 过滤）/ 覆盖 / 更新（upsert）/
> 并发分片 / 自动建表。
>
> **连接信息不通过接口传递**：来源与目标的 MySQL 连接、表名一律取自已保存的
> **采集（Collect）/ 同步（Sync）节点 content**（请求传 `node_id`），接口只接收运行参数。

## 一、节点 content 与连接解析

### 1. Sync 节点（推荐：任意两端 MySQL）

`SyncNodeContent`：

```json
{
  "source": {
    "datasource_id": "ds-src",
    "table": "orders",
    "database": "srcdb",
    "columns": [
      { "name": "id", "data_type": "bigint", "is_primary": 1 }
    ],
    "custom_sql": ""
  },
  "target": {
    "datasource_id": "ds-dst",
    "table": "dwd_orders",
    "database": "dstdb",
    "create_ddl": "",
    "is_cover": 1
  }
}
```

| 字段 | 说明 |
|---|---|
| source.datasource_id | 源数据源 ID（租户库 `datasource` 表；`OdsDatasourceId` 指公共数据源） |
| source.table | 源表名（必填） |
| source.database | 可选，覆盖数据源自带库名 |
| source.columns | 可选字段子集；空 = 全列 |
| target.datasource_id | 目标数据源 ID（必填） |
| target.table | 目标表名（必填） |
| target.database | 可选，覆盖数据源自带库名 |
| target.create_ddl | 可选自定义建表 DDL；非空时目标表缺失默认自动建表 |
| target.is_cover | 1 = 覆盖写入（默认 mode=replace）；接口显式传 mode 时以接口为准 |

### 2. Collect 节点（源 = 外部数据源，目标 = 当前租户库）

`CollectNodeContent`（现有结构不变）：

```json
{
  "source": { "datasource_id": "ds-ext", "table": "src_user" },
  "target": { "table": "dwd_user", "create_ddl": "", "is_cover": 0 }
}
```

| 字段 | 说明 |
|---|---|
| source.datasource_id | 源数据源 ID（外部库，必填） |
| source.table | 源表名（必填） |
| target.table | 目标表名（必填） |
| target.create_ddl | 可选自定义建表 DDL |
| target.is_cover | 1 = 默认 mode=replace |

目标连接 = **当前租户库**（服务全局 MySQL 连接参数 + 租户库名），无需配置数据源。

## 二、接口

### 1. 提交迁移任务 `POST /api/migrate/run`

请求体：

```json
{
  "run_id": "mig-20260918-01",
  "node_id": "node-sync-0001",
  "where": "id > 1000",
  "batch_size": 2000,
  "mode": "upsert",
  "create_table_if_missing": true,
  "channels": 4
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| run_id | string | 必填，幂等键。运行中重复提交返回 400；已结束任务可用同 run_id 覆盖重跑 |
| node_id | string | 必填，Collect/Sync 节点 ID，连接与表名取自其 content |
| where | string | 可选过滤（原样拼进 SELECT，勿带 WHERE 关键字） |
| batch_size | int | 每事务写入行数，默认 2000 |
| mode | string | 可选：insert / replace / upsert；缺省取节点 is_cover（1→replace，否则 insert） |
| create_table_if_missing | bool | 可选。true=目标表缺失自动建表（节点有 create_ddl 优先用自定义 DDL，否则复制源表 DDL）；false=关闭；缺省=节点是否配置 create_ddl |
| channels | int | 可选并发通道数。>1 时按源表数字主键切分区间并行，无数字主键自动回退串行 |

响应：`{"code":200,"data":"ok"}`；绑定/解析错误返回 400，任务已存在返回 400。

### 2. 查询状态 `GET /api/migrate/status?run_id=<id>`

```json
{
  "code": 200,
  "data": {
    "run_id": "mig-20260918-01",
    "status": "success",
    "rows_read": 5000,
    "rows_written": 5000,
    "batches": 12,
    "message": "迁移完成",
    "started_at": "2026-09-18T10:07:00+08:00",
    "finished_at": "2026-09-18T10:07:00+08:00"
  }
}
```

status：`running / success / failed / canceled`；任务不存在返回 404。

### 3. 取消任务 `POST /api/migrate/stop?run_id=<id>`

取消运行中任务（在批次边界生效，不产生半批脏数据）。任务不存在返回 404，已结束返回 400。

## 三、行为约定与边界

- 仅需登录（JWT），不依赖租户中间件；节点解析时后端自行取租户库名与数据源。
- 任务状态存内存，进程重启后记录丢失（如需持久化可后续落库）。
- channels 并发依赖源表**单列数字主键**；复合主键/无主键回退串行。
- 每批写入单事务，失败整批回滚。
- where 需保证 SQL 合法（与 DataX 语义一致）。
