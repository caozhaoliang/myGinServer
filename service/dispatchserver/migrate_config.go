package dispatchserver

import (
	"context"
	"encoding/json"

	mdispatch "myGinServer/models/dispatch"
	"myGinServer/pkg/tenantctx"
	"myGinServer/service/migrate"

	"github.com/pkg/errors"
)

// MigrateConfigFromNode 从已保存的采集/同步节点 content 解析出迁移配置，
// 连接信息一律来自节点配置，接口不再传递来源与目标的连接信息：
//
//   - Sync 节点（SyncNodeContent）：源/目标均为数据源引用（datasource_id），
//     支持任意两端 MySQL；表名分别取 source.table / target.table。
//   - Collect 节点（CollectNodeContent）：源 = source.datasource_id（外部库），
//     目标 = 当前租户库（服务自身 MySQL 连接 + 租户库名）；表名取 source.table / target.table。
//
// 运行参数（where/batch_size/mode/channels 等）不属于节点配置，由调用方从请求覆盖。
func (n *NodeServer) MigrateConfigFromNode(ctx context.Context, nodeID string) (*migrate.Config, error) {
	node, err := n.store.GetNode(ctx, tenantctx.TenantDB(ctx), nodeID)
	if err != nil {
		return nil, errors.Wrapf(err, "获取节点失败: %s", nodeID)
	}
	switch mdispatch.NodeType(node.Type) {
	case mdispatch.NodeSync:
		var c mdispatch.SyncNodeContent
		if err := json.Unmarshal([]byte(node.Content.String), &c); err != nil {
			return nil, errors.Wrap(err, "解析同步节点 content 失败")
		}
		if c.Source.DatasourceId == "" || c.Target.DatasourceId == "" {
			return nil, errors.New("同步节点 source/target 均需配置数据源 ID")
		}
		if c.Source.Table == "" || c.Target.Table == "" {
			return nil, errors.New("同步节点 source/target 均需配置表名")
		}
		src, err := n.resolveDatasource(ctx, c.Source.DatasourceId)
		if err != nil {
			return nil, errors.Wrap(err, "解析源数据源失败")
		}
		tgt, err := n.resolveDatasource(ctx, c.Target.DatasourceId)
		if err != nil {
			return nil, errors.Wrap(err, "解析目标数据源失败")
		}
		if c.Source.Database != "" {
			src.Database = c.Source.Database
		}
		if c.Target.Database != "" {
			tgt.Database = c.Target.Database
		}
		// is_cover=1 表示覆盖写入：未显式指定 mode 时默认 replace
		mode := ""
		if c.Target.IsCover == 1 {
			mode = "replace"
		}
		return &migrate.Config{
			Source:               toMigrateEndpoint(src),
			Target:               toMigrateEndpoint(tgt),
			Table:                c.Source.Table,
			TargetTable:          c.Target.Table,
			Columns:              columnNames(c.Source.Columns),
			Mode:                 mode,
			CreateDDL:            c.Target.CreateDDL,
			CreateTableIfMissing: c.Target.CreateDDL != "",
		}, nil

	case mdispatch.NodeCollect:
		var c mdispatch.CollectNodeContent
		if err := json.Unmarshal([]byte(node.Content.String), &c); err != nil {
			return nil, errors.Wrap(err, "解析采集节点 content 失败")
		}
		if c.Source.DatasourceId == "" {
			return nil, errors.New("采集节点 source 需配置数据源 ID")
		}
		if c.Source.Table == "" || c.Target.Table == "" {
			return nil, errors.New("采集节点 source/target 均需配置表名")
		}
		src, err := n.resolveDatasource(ctx, c.Source.DatasourceId)
		if err != nil {
			return nil, errors.Wrap(err, "解析源数据源失败")
		}
		if c.Source.Database != "" {
			src.Database = c.Source.Database
		}
		tgt, err := n.tenantDBEndpoint(ctx)
		if err != nil {
			return nil, err
		}
		mode := ""
		if c.Target.IsCover == 1 {
			mode = "replace"
		}
		return &migrate.Config{
			Source:               toMigrateEndpoint(src),
			Target:               toMigrateEndpoint(tgt),
			Table:                c.Source.Table,
			TargetTable:          c.Target.Table,
			Columns:              columnNames(c.Source.Columns),
			Mode:                 mode,
			CreateDDL:            c.Target.CreateDDL,
			CreateTableIfMissing: c.Target.CreateDDL != "",
		}, nil

	default:
		return nil, errors.Errorf("节点类型 %s 不支持数据迁移（仅支持 Collect/Sync）", node.Type)
	}
}

// columnNames 提取节点 content 中配置的字段名清单（为空表示全列）。
func columnNames(cols []mdispatch.Column) []string {
	if len(cols) == 0 {
		return nil
	}
	names := make([]string, 0, len(cols))
	for _, c := range cols {
		if c.Name != "" {
			names = append(names, c.Name)
		}
	}
	return names
}

// resolveDatasource 按数据源 ID 解析连接信息：优先租户库内的数据源，
// 固定 ODS 公共数据源走公共库（project 为空）。
func (n *NodeServer) resolveDatasource(ctx context.Context, dsId string) (*mdispatch.MysqlDatasourceContent, error) {
	project := tenantctx.TenantDB(ctx)
	if dsId == OdsDatasourceId {
		project = ""
	}
	return n.getDatasourceContent(ctx, project, dsId)
}

// tenantDBEndpoint 构造当前租户库（项目库）的连接信息：连接参数来自服务全局配置，
// 库名取租户上下文（缺失时回退默认库）。
func (n *NodeServer) tenantDBEndpoint(ctx context.Context) (*mdispatch.MysqlDatasourceContent, error) {
	dbName := n.conf.DBConfig.DbName
	if t := tenantctx.TenantDB(ctx); t != "" {
		dbName = t
	}
	return &mdispatch.MysqlDatasourceContent{
		User:     n.conf.DBConfig.DbUser,
		Passwd:   n.conf.DBConfig.DbPassword,
		Host:     n.conf.DBConfig.DbHost,
		Port:     n.conf.DBConfig.DbPort,
		Database: dbName,
	}, nil
}

// toMigrateEndpoint 把数据源连接信息转为迁移引擎端点。
func toMigrateEndpoint(c *mdispatch.MysqlDatasourceContent) migrate.Endpoint {
	return migrate.Endpoint{
		Host:     c.Host,
		Port:     c.Port,
		User:     c.User,
		Password: c.Passwd,
		Database: c.Database,
	}
}
