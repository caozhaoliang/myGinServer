package utils

import (
	"fmt"
	"testing"
)

func TestSQLRender(t *testing.T) {
	// 值：自动加引号
	out, _ := RenderTemplate(
		"select {{key}} from dual",
		map[string]interface{}{"key": "val"},
	)
	fmt.Println(out) // select 'val' from dual

	// 数字：不加引号
	out, _ = RenderTemplate(
		"select {{n}} from dual",
		map[string]interface{}{"n": 123},
	)
	fmt.Println(out) // select 123 from dual

	// 标识符：反引号包裹
	out, _ = RenderTemplate(
		"select * from {{table}} where {{col}} = {{val}}",
		map[string]interface{}{
			"table": Identifier("users"),
			"col":   Identifier("name"),
			"val":   "张三",
		},
	)
	fmt.Println(out) // select * from `users` where `name` = '张三'
}
