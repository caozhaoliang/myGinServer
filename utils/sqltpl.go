package utils

import (
	"fmt"
	"strings"
)

// RenderTemplate 把 SQL 里的 {{key}} 替换成 params 中对应的值。
// 规则：
//   - 值类型（string/int/bool/...）会格式化成 SQL 字面量，string 自动加单引号
//   - 值是 Identifier 类型时，用反引号包裹，用于表名/列名
//   - 找不到 key 直接报错，避免静默生成错误 SQL
func RenderTemplate(sql string, params map[string]interface{}) (string, error) {
	var b strings.Builder
	b.Grow(len(sql) + 64)

	i := 0
	n := len(sql)

	for i < n {
		// 找下一个 {{
		start := strings.Index(sql[i:], "{{")
		if start == -1 {
			b.WriteString(sql[i:])
			break
		}
		start += i

		// {{ 之前的内容原样输出
		b.WriteString(sql[i:start])

		// 找对应的 }}
		end := strings.Index(sql[start+2:], "}}")
		if end == -1 {
			// 没有闭合，原样输出剩余内容
			b.WriteString(sql[start:])
			break
		}
		end = start + 2 + end

		key := strings.TrimSpace(sql[start+2 : end])
		if key == "" {
			return "", fmt.Errorf("空的占位符 {{}}")
		}

		val, ok := params[key]
		if !ok {
			return "", fmt.Errorf("参数缺失: %s", key)
		}

		s, err := formatValue(val)
		if err != nil {
			return "", fmt.Errorf("参数 %s: %w", key, err)
		}
		b.WriteString(s)

		i = end + 2
	}

	return b.String(), nil
}

// Identifier 用于标记“这是表名/列名，不是值”
type Identifier string

// formatValue 把参数值转成 SQL 片段
func formatValue(v interface{}) (string, error) {
	switch x := v.(type) {
	case nil:
		return "NULL", nil
	case Identifier:
		return "`" + strings.ReplaceAll(string(x), "`", "``") + "`", nil
	case string:
		return "'" + escapeString(x) + "'", nil
	case []byte:
		return "'" + escapeString(string(x)) + "'", nil
	case bool:
		if x {
			return "1", nil
		}
		return "0", nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return fmt.Sprintf("%v", x), nil
	default:
		return "", fmt.Errorf("不支持的类型: %T", v)
	}
}

// escapeString 转义 MySQL 字符串字面量
func escapeString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			b.WriteString("''")
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case 0:
			b.WriteString("\\0")
		case 0x1a:
			b.WriteString("\\Z")
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
