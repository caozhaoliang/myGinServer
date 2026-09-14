package dispatchserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"myGinServer/api/request"
	"myGinServer/api/response"
	mdispatch "myGinServer/models/dispatch"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func (n *NodeServer) SaveDatasource(ctx context.Context, req request.DatasourceReq) error {
	err := n.store.SaveDatasource(ctx, "", mdispatch.Datasource{
		Id:        req.Id,
		Code:      req.Code,
		Name:      req.Name,
		Type:      req.Type,
		ConnStr:   req.ConnStr,
		CreatedOn: sql.NullTime{Time: time.Now(), Valid: true},
		CreatedBy: sql.NullString{String: "admin", Valid: true},
	})

	return err
}

func (n *NodeServer) ListDatasource(ctx context.Context) ([]request.DatasourceReq, error) {
	datasource, err := n.store.ListDatasource(ctx, "")
	if err != nil {
		return nil, err
	}
	var res []request.DatasourceReq
	for _, ds := range datasource {
		res = append(res, request.DatasourceReq{
			Id:      ds.Id,
			Code:    ds.Code,
			Name:    ds.Name,
			Type:    ds.Type,
			ConnStr: ds.ConnStr,
		})
	}
	return res, nil
}

// MetaTables 根据数据源 ID 获取该数据源下所有表的列表
func (n *NodeServer) MetaTables(ctx context.Context, dsId string) ([]response.MetaTables, error) {
	// 根据数据源 ID 获取数据源连接信息
	content, err := n.getDatasourceContent(ctx, dsId)
	if err != nil {
		return nil, err
	}
	// 打开目标数据库连接
	db, err := openTargetDB(content)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_NAME, COALESCE(TABLE_COMMENT, '')
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME`, content.Database)
	if err != nil {
		return nil, errors.Wrap(err, "查询表列表失败")
	}
	defer rows.Close()

	var tables []response.MetaTables
	for rows.Next() {
		var t response.MetaTables
		if err := rows.Scan(&t.Name, &t.Description); err != nil {
			return nil, errors.Wrap(err, "读取表信息失败")
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "遍历表列表失败")
	}
	return tables, nil
}

// MetaColumns 根据数据源 ID 与表名获取该表下的所有列信息
func (n *NodeServer) MetaColumns(ctx context.Context, dsId, table string) ([]response.MetaColumns, error) {
	// 根据数据源 ID 获取数据源连接信息
	content, err := n.getDatasourceContent(ctx, dsId)
	if err != nil {
		return nil, err
	}
	// 打开目标数据库连接
	db, err := openTargetDB(content)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT c.COLUMN_NAME,
		       c.COLUMN_TYPE,
		       COALESCE(c.COLUMN_COMMENT, ''),
		       COALESCE(k.ORDINAL_POSITION, 0)
		FROM information_schema.COLUMNS c
		LEFT JOIN information_schema.KEY_COLUMN_USAGE k
		       ON k.TABLE_SCHEMA = c.TABLE_SCHEMA
		      AND k.TABLE_NAME = c.TABLE_NAME
		      AND k.COLUMN_NAME = c.COLUMN_NAME
		      AND k.CONSTRAINT_NAME = 'PRIMARY'
		WHERE c.TABLE_SCHEMA = ? AND c.TABLE_NAME = ?
		ORDER BY c.ORDINAL_POSITION`, content.Database, table)
	if err != nil {
		return nil, errors.Wrap(err, "查询列信息失败")
	}
	defer rows.Close()

	var columns []response.MetaColumns
	for rows.Next() {
		var col response.MetaColumns
		if err := rows.Scan(&col.Name, &col.Type, &col.Comment, &col.PrimaryKeySeq); err != nil {
			return nil, errors.Wrap(err, "读取列信息失败")
		}
		columns = append(columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "遍历列信息失败")
	}
	return columns, nil
}

// getDatasourceContent 根据数据源 ID 获取并解析连接信息，目前仅支持 mysql 类型
func (n *NodeServer) getDatasourceContent(ctx context.Context, dsId string) (*mdispatch.MysqlDatasourceContent, error) {
	ds, err := n.store.GetDatasource(ctx, "", dsId)
	if err != nil {
		return nil, errors.Wrapf(err, "获取数据源失败: %s", dsId)
	}
	if ds.Type != "" && ds.Type != "mysql" {
		return nil, errors.Errorf("暂不支持的数据源类型: %s", ds.Type)
	}
	var content mdispatch.MysqlDatasourceContent
	if err := json.Unmarshal([]byte(ds.ConnStr), &content); err != nil {
		return nil, errors.Wrap(err, "解析连接信息失败")
	}
	if content.Database == "" {
		content.Database = content.Schema
	}
	return &content, nil
}

// openTargetDB 根据连接信息打开目标 MySQL 数据库连接
func openTargetDB(content *mdispatch.MysqlDatasourceContent) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = content.User
	cfg.Passwd = content.Passwd
	cfg.Addr = fmt.Sprintf("%s:%d", content.Host, content.Port)
	cfg.DBName = content.Database
	cfg.Net = "tcp"
	cfg.Params = map[string]string{
		"charset":   "utf8mb4", // 字符集
		"parseTime": "true",    // 自动处理时间类型
		"loc":       "Local",   // 时区
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, errors.Wrap(err, "打开目标数据库连接失败")
	}
	return db, nil
}
