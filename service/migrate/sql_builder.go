package migrate

import (
	"fmt"
	"strings"
)

// QuoteIdent 用反引号包裹标识符（表名/列名），内部反引号按 MySQL 规则双写转义。
func QuoteIdent(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

// buildSelect 组装源表查询：SELECT `c1`,`c2` FROM `tbl` [WHERE <where>]
// where 为过滤条件（不含 WHERE 关键字），为空则不带过滤。
func buildSelect(table string, columns []string, where string) (string, error) {
	if table == "" || len(columns) == 0 {
		return "", fmt.Errorf("表名与列清单不能为空")
	}
	var b strings.Builder
	b.WriteString("SELECT ")
	for i, c := range columns {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(QuoteIdent(c))
	}
	b.WriteString(" FROM ")
	b.WriteString(QuoteIdent(table))
	if strings.TrimSpace(where) != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}
	return b.String(), nil
}

// buildInsert 组装批量写入语句。
// mode：insert=INSERT INTO；replace=REPLACE INTO；upsert=INSERT ... ON DUPLICATE KEY UPDATE。
// 行数由调用方保证与 placeholders 匹配。
func buildInsert(table string, columns []string, rows int, mode string) (string, error) {
	if table == "" || len(columns) == 0 || rows <= 0 {
		return "", fmt.Errorf("表名、列清单、行数不能为空")
	}
	cols := make([]string, len(columns))
	for i, c := range columns {
		cols[i] = QuoteIdent(c)
	}

	var b strings.Builder
	switch mode {
	case "replace":
		b.WriteString("REPLACE INTO ")
	case "upsert":
		b.WriteString("INSERT INTO ")
	default:
		b.WriteString("INSERT INTO ")
	}
	b.WriteString(QuoteIdent(table))
	b.WriteString(" (")
	b.WriteString(strings.Join(cols, ","))
	b.WriteString(") VALUES ")

	placeholders := "(" + strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",") + ")"
	rowStmts := make([]string, rows)
	for i := 0; i < rows; i++ {
		rowStmts[i] = placeholders
	}
	b.WriteString(strings.Join(rowStmts, ","))

	if mode == "upsert" {
		updates := make([]string, len(cols))
		for i, c := range cols {
			updates[i] = c + "=VALUES(" + c + ")"
		}
		b.WriteString(" ON DUPLICATE KEY UPDATE ")
		b.WriteString(strings.Join(updates, ","))
	}
	return b.String(), nil
}

// ---- 主键探测与并发分片 ----

// pkProbeSQL 取表的单列主键名与数据类型（按 ORDINAL_POSITION 排序，单列主键时取第一条）。
const pkProbeSQL = `SELECT k.COLUMN_NAME, c.DATA_TYPE
	FROM information_schema.KEY_COLUMN_USAGE k
	JOIN information_schema.COLUMNS c
	  ON c.TABLE_SCHEMA = k.TABLE_SCHEMA AND c.TABLE_NAME = k.TABLE_NAME AND c.COLUMN_NAME = k.COLUMN_NAME
	WHERE k.TABLE_SCHEMA = ? AND k.TABLE_NAME = ? AND k.CONSTRAINT_NAME = 'PRIMARY'
	ORDER BY k.ORDINAL_POSITION`

// isNumericPKType 判断主键列是否为可安全做 BETWEEN 区间切分的整数类型。
func isNumericPKType(dataType string) bool {
	switch strings.ToLower(dataType) {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint", "year":
		return true
	}
	return false
}

// pkRange 一个并发通道负责的主键区间（闭区间 [Lo, Hi]）。
type pkRange struct {
	Lo int64
	Hi int64
}

// splitPKRanges 把 [min, max] 按 channels 切成尽量均匀的闭区间。
// 最后一段收口到 max，避免浮点溢出；channels<=1 返回单个全量区间。
func splitPKRanges(minV, maxV int64, channels int) []pkRange {
	if channels <= 1 {
		return []pkRange{{Lo: minV, Hi: maxV}}
	}
	if minV > maxV {
		return nil
	}
	if minV == maxV {
		return []pkRange{{Lo: minV, Hi: maxV}}
	}
	span := maxV - minV + 1
	step := span / int64(channels)
	if step < 1 {
		step = 1
	}
	ranges := make([]pkRange, 0, channels)
	lo := minV
	for i := 0; i < channels; i++ {
		hi := lo + step - 1
		if i == channels-1 {
			hi = maxV
		}
		if hi > maxV {
			hi = maxV
		}
		ranges = append(ranges, pkRange{Lo: lo, Hi: hi})
		lo = hi + 1
		if lo > maxV {
			break
		}
	}
	return ranges
}
