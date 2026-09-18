package migrate

import (
	"strings"
	"testing"
)

func TestQuoteIdent(t *testing.T) {
	cases := map[string]string{
		"user":     "`user`",
		"weird`id": "`weird``id`",
	}
	for in, want := range cases {
		if got := QuoteIdent(in); got != want {
			t.Fatalf("QuoteIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildSelect(t *testing.T) {
	cols := []string{"id", "name", "create_time"}
	q, err := buildSelect("user", cols, "")
	if err != nil {
		t.Fatal(err)
	}
	want := "SELECT `id`,`name`,`create_time` FROM `user`"
	if q != want {
		t.Fatalf("无 where: got %q, want %q", q, want)
	}

	q, _ = buildSelect("user", cols, "id > 100 AND status = 'ok'")
	want = "SELECT `id`,`name`,`create_time` FROM `user` WHERE id > 100 AND status = 'ok'"
	if q != want {
		t.Fatalf("带 where: got %q, want %q", q, want)
	}

	if _, err := buildSelect("", cols, ""); err == nil {
		t.Fatal("空表名应报错")
	}
	if _, err := buildSelect("t", nil, ""); err == nil {
		t.Fatal("空列清单应报错")
	}
}

func TestBuildInsert(t *testing.T) {
	cols := []string{"id", "name"}

	for _, mode := range []string{"insert", "replace", "upsert"} {
		sqlStr, err := buildInsert("tbl", cols, 2, mode)
		if err != nil {
			t.Fatal(err)
		}
		switch mode {
		case "insert":
			if !strings.HasPrefix(sqlStr, "INSERT INTO `tbl`") {
				t.Fatalf("insert 前缀错误: %s", sqlStr)
			}
		case "replace":
			if !strings.HasPrefix(sqlStr, "REPLACE INTO `tbl`") {
				t.Fatalf("replace 前缀错误: %s", sqlStr)
			}
		case "upsert":
			if !strings.Contains(sqlStr, "ON DUPLICATE KEY UPDATE") ||
				!strings.Contains(sqlStr, "`id`=VALUES(`id`)") {
				t.Fatalf("upsert 语句不完整: %s", sqlStr)
			}
		}
		// 占位符数量 = 列数 × 行数
		if n := strings.Count(sqlStr, "?"); n != 4 {
			t.Fatalf("mode=%s 占位符数量 %d, want 4", mode, n)
		}
	}

	if _, err := buildInsert("t", cols, 0, "insert"); err == nil {
		t.Fatal("行数 0 应报错")
	}
	// 非法 mode 按 insert 处理（绑定层已 oneof 校验）
	sqlStr, _ := buildInsert("t", cols, 1, "whatever")
	if !strings.HasPrefix(sqlStr, "INSERT INTO") {
		t.Fatalf("未知 mode 应回退 insert: %s", sqlStr)
	}
}

func TestSplitPKRanges(t *testing.T) {
	// 单通道：全量区间
	rs := splitPKRanges(1, 100, 1)
	if len(rs) != 1 || rs[0].Lo != 1 || rs[0].Hi != 100 {
		t.Fatalf("单通道区间错误: %+v", rs)
	}
	// 3 通道均匀切分：100 行分 3 段，最后一段收口到 max
	rs = splitPKRanges(1, 100, 3)
	if len(rs) != 3 {
		t.Fatalf("3 通道区间数错误: %+v", rs)
	}
	if rs[0].Lo != 1 || rs[0].Hi != 33 {
		t.Fatalf("第一段错误: %+v", rs[0])
	}
	if rs[2].Lo != 67 || rs[2].Hi != 100 {
		t.Fatalf("最后一段未收口到 max: %+v", rs[2])
	}
	// 覆盖性：区间无缝衔接 [1,100]
	prevHi := int64(0)
	for _, r := range rs {
		if r.Lo != prevHi+1 {
			t.Fatalf("区间断裂: %+v", rs)
		}
		prevHi = r.Hi
	}
	if prevHi != 100 {
		t.Fatalf("区间未覆盖到 100: %+v", rs)
	}
	// min == max：单行
	rs = splitPKRanges(7, 7, 4)
	if len(rs) != 1 || rs[0].Lo != 7 || rs[0].Hi != 7 {
		t.Fatalf("单行区间错误: %+v", rs)
	}
	// span 小于 channels：退化为每段至少 1，段数不超过行数
	rs = splitPKRanges(1, 3, 5)
	if len(rs) > 3 {
		t.Fatalf("段数不应超过行数: %+v", rs)
	}
	// 空集
	if rs := splitPKRanges(5, 1, 2); rs != nil {
		t.Fatalf("min>max 应为空: %+v", rs)
	}
}

func TestIsNumericPKType(t *testing.T) {
	for _, ok := range []string{"int", "bigint", "INT", "smallint", "mediumint", "tinyint", "year"} {
		if !isNumericPKType(ok) {
			t.Fatalf("%s 应为数字主键类型", ok)
		}
	}
	for _, no := range []string{"varchar", "char", "datetime", "decimal", "float", "double", "text"} {
		if isNumericPKType(no) {
			t.Fatalf("%s 不应是数字主键类型", no)
		}
	}
}

func TestPKProbeSQL(t *testing.T) {
	if !strings.Contains(pkProbeSQL, "information_schema.KEY_COLUMN_USAGE") {
		t.Fatal("主键探测 SQL 应查 information_schema")
	}
}
