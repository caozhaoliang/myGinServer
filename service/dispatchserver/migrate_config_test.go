package dispatchserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"myGinServer/config"
	"myGinServer/internal/store/dispatch"
	mdispatch "myGinServer/models/dispatch"
	"myGinServer/pkg/tenantctx"

	"github.com/pkg/errors"
)

// stubStore 嵌入接口以便只覆写 GetNode/GetDatasource，同时记录公共库访问。
type stubStore struct {
	dispatch.StoreIface
	node *mdispatch.Nodes
	ds   map[string]mdispatch.Datasource
	// 记录 getDatasourceContent 走的是哪个 project（"" 表示公共库）
	projects []string
}

func (s *stubStore) GetNode(_ context.Context, _ string, _ string) (*mdispatch.Nodes, error) {
	if s.node == nil {
		return nil, errors.New("节点不存在")
	}
	return s.node, nil
}

func (s *stubStore) GetDatasource(_ context.Context, project, id string) (mdispatch.Datasource, error) {
	s.projects = append(s.projects, project)
	ds, ok := s.ds[id]
	if !ok {
		return mdispatch.Datasource{}, errors.Errorf("数据源不存在: %s", id)
	}
	return ds, nil
}

func testConf() *config.Config {
	return &config.Config{DBConfig: config.DBConfig{
		DbHost: "127.0.0.1", DbPort: 3306, DbUser: "root", DbPassword: "12345678", DbName: "mytest",
	}}
}

func testCtx() context.Context {
	ctx := tenantctx.WithTenantDB(context.Background(), "mytest")
	return tenantctx.WithUserID(ctx, "u1")
}

func dsContent(user, passwd string, port int, database string) string {
	b, _ := json.Marshal(mdispatch.MysqlDatasourceContent{
		User: user, Passwd: passwd, Host: "10.0.0.8", Port: port, Database: database,
	})
	return string(b)
}

func newNodeServer(t *testing.T, node *mdispatch.Nodes, ds map[string]mdispatch.Datasource) (*NodeServer, *stubStore) {
	t.Helper()
	st := &stubStore{node: node, ds: ds}
	ns := NewNodeServer(st, nil, testConf())
	return ns, st
}

func TestMigrateConfigFromNode_Sync(t *testing.T) {
	content, _ := json.Marshal(mdispatch.SyncNodeContent{
		Source: mdispatch.CollectSource{
			DatasourceId: "ds-src", Table: "orders",
			Columns: []mdispatch.Column{{Name: "id"}, {Name: "name"}},
		},
		Target: mdispatch.SyncTarget{
			DatasourceId: "ds-dst", Table: "dwd_orders", Database: "odstarget", IsCover: 1,
			CreateDDL: "CREATE TABLE `dwd_orders` (id INT)",
		},
	})
	node := &mdispatch.Nodes{Id: "n1", Type: mdispatch.NodeSync, Content: sql.NullString{String: string(content), Valid: true}}
	ns, st := newNodeServer(t, node, map[string]mdispatch.Datasource{
		"ds-src": {Id: "ds-src", Type: "mysql", ConnStr: dsContent("src", "sp", 3306, "srcdb")},
		"ds-dst": {Id: "ds-dst", Type: "mysql", ConnStr: dsContent("dst", "dp", 3307, "dstdb")},
	})

	cfg, err := ns.MigrateConfigFromNode(testCtx(), "n1")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source.Host != "10.0.0.8" || cfg.Source.Port != 3306 || cfg.Source.User != "src" || cfg.Source.Database != "srcdb" {
		t.Fatalf("源端点解析错误: %+v", cfg.Source)
	}
	if cfg.Target.Database != "odstarget" { // 节点 database 覆盖数据源自带库
		t.Fatalf("目标库未被节点 database 覆盖: %+v", cfg.Target)
	}
	if cfg.Table != "orders" || cfg.TargetTable != "dwd_orders" {
		t.Fatalf("表名解析错误: %s / %s", cfg.Table, cfg.TargetTable)
	}
	if cfg.Mode != "replace" { // is_cover=1 → replace
		t.Fatalf("mode 应为 replace, got %q", cfg.Mode)
	}
	if cfg.CreateDDL == "" || !cfg.CreateTableIfMissing {
		t.Fatalf("create_ddl 应默认启用建表: %+v", cfg)
	}
	if len(cfg.Columns) != 2 || cfg.Columns[0] != "id" || cfg.Columns[1] != "name" {
		t.Fatalf("列提取错误: %v", cfg.Columns)
	}
	// 两端数据源都从租户库解析（非 ODS）
	for _, p := range st.projects {
		if p != "mytest" {
			t.Fatalf("数据源应从租户库解析, got project=%q", p)
		}
	}
}

func TestMigrateConfigFromNode_Sync_ODS(t *testing.T) {
	content, _ := json.Marshal(mdispatch.SyncNodeContent{
		Source: mdispatch.CollectSource{DatasourceId: OdsDatasourceId, Table: "ods_tbl"},
		Target: mdispatch.SyncTarget{DatasourceId: "ds-dst", Table: "dst_tbl"},
	})
	node := &mdispatch.Nodes{Id: "n2", Type: mdispatch.NodeSync, Content: sql.NullString{String: string(content), Valid: true}}
	ns, st := newNodeServer(t, node, map[string]mdispatch.Datasource{
		OdsDatasourceId: {Id: OdsDatasourceId, Type: "mysql", ConnStr: dsContent("ods", "o", 3306, "odsdb")},
		"ds-dst":        {Id: "ds-dst", Type: "mysql", ConnStr: dsContent("dst", "d", 3306, "dstdb")},
	})

	cfg, err := ns.MigrateConfigFromNode(testCtx(), "n2")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source.Database != "odsdb" || cfg.Source.User != "ods" {
		t.Fatalf("ODS 数据源解析错误: %+v", cfg.Source)
	}
	// 源走公共库（project=""），目标走租户库
	if st.projects[0] != "" || st.projects[1] != "mytest" {
		t.Fatalf("ODS 应走公共库: %v", st.projects)
	}
}

func TestMigrateConfigFromNode_Collect(t *testing.T) {
	content, _ := json.Marshal(mdispatch.CollectNodeContent{
		Source: mdispatch.CollectSource{DatasourceId: "ds-ext", Table: "ext_user", Database: "extdb"},
		Target: mdispatch.CollectTarget{Table: "dwd_user"},
	})
	node := &mdispatch.Nodes{Id: "n3", Type: mdispatch.NodeCollect, Content: sql.NullString{String: string(content), Valid: true}}
	ns, _ := newNodeServer(t, node, map[string]mdispatch.Datasource{
		"ds-ext": {Id: "ds-ext", Type: "mysql", ConnStr: dsContent("ext", "e", 3306, "extdb")},
	})

	cfg, err := ns.MigrateConfigFromNode(testCtx(), "n3")
	if err != nil {
		t.Fatal(err)
	}
	// 源 = 外部数据源
	if cfg.Source.User != "ext" || cfg.Source.Database != "extdb" {
		t.Fatalf("源端点错误: %+v", cfg.Source)
	}
	// 目标 = 当前租户库（全局连接 + tenantctx 库名）
	if cfg.Target.Host != "127.0.0.1" || cfg.Target.User != "root" || cfg.Target.Database != "mytest" {
		t.Fatalf("目标应为租户库: %+v", cfg.Target)
	}
	if cfg.Table != "ext_user" || cfg.TargetTable != "dwd_user" {
		t.Fatalf("表名错误: %s / %s", cfg.Table, cfg.TargetTable)
	}
	if cfg.Mode != "" { // is_cover=0 → 不默认 replace
		t.Fatalf("is_cover=0 时 mode 应为空, got %q", cfg.Mode)
	}
}

func TestMigrateConfigFromNode_Errors(t *testing.T) {
	// 不支持的类型
	node := &mdispatch.Nodes{Id: "n4", Type: mdispatch.NodeSQL, Content: sql.NullString{String: `{"sql":"select 1"}`, Valid: true}}
	ns, _ := newNodeServer(t, node, nil)
	if _, err := ns.MigrateConfigFromNode(testCtx(), "n4"); err == nil {
		t.Fatal("SQL 节点应报不支持")
	}

	// Sync 缺数据源 ID
	content, _ := json.Marshal(mdispatch.SyncNodeContent{
		Source: mdispatch.CollectSource{Table: "t"},
		Target: mdispatch.SyncTarget{DatasourceId: "d", Table: "t2"},
	})
	node = &mdispatch.Nodes{Id: "n5", Type: mdispatch.NodeSync, Content: sql.NullString{String: string(content), Valid: true}}
	ns, _ = newNodeServer(t, node, nil)
	if _, err := ns.MigrateConfigFromNode(testCtx(), "n5"); err == nil {
		t.Fatal("缺数据源 ID 应报错")
	}

	// content 非法 JSON
	node = &mdispatch.Nodes{Id: "n6", Type: mdispatch.NodeSync, Content: sql.NullString{String: "not-json", Valid: true}}
	ns, _ = newNodeServer(t, node, nil)
	if _, err := ns.MigrateConfigFromNode(testCtx(), "n6"); err == nil {
		t.Fatal("非法 JSON 应报错")
	}
}
