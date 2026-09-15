package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestNewCronParse(t *testing.T) {
	// 示例：找出 2026年9月13日 当天，每周一到周五 9:30 的所有触发时间
	// 注意：ParseStandard 只接受标准 5 字段表达式
	spec := "30 9 * * 1-5"

	// 定义时间段：从当天 00:00 到第二天 00:00
	loc := time.Local // 或者指定时区，如 time.LoadLocation("Asia/Shanghai")
	start := time.Date(2026, 9, 13, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 2) // 加2天

	times, err := GetSchedulesBetween(spec, start, end)
	if err != nil {
		panic(err)
	}

	fmt.Printf("在 %s 到 %s 之间，表达式 '%s' 触发了 %d 次：\n", start, end, spec, len(times))
	for _, t := range times {
		fmt.Println(" -", t.Format("2006-01-02 15:04:05"))
	}
}

// TestGetSchedulesBetween_Boundary 回归测试：窗口左端必须为闭区间。
// cron 的 Next 返回的是「严格大于」入参的时间点，若直接用 start 起步，
// 恰好落在 start 上的触发（如 0 0 * * * 落在窗口起点的午夜整点）会被整条吞掉，
// 表现为根节点在批次窗口内生成 0 个实例。
// 这里统一用 UTC，避免宿主机夏令时导致「一天不是 1440 分钟」而干扰断言。
func TestGetSchedulesBetween_Boundary(t *testing.T) {
	// 业务日期 2026-09-15 对应的调度窗口为 2026-09-16（周三）整日
	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	cases := []struct {
		name     string
		spec     string
		wantLen  int
		wantHead string // wantLen 为 0 时忽略
	}{
		{"整点表达式命中窗口左端", "0 0 * * *", 1, "2026-09-16 00:00:00"},
		{"日间表达式正常命中", "30 2 * * *", 1, "2026-09-16 02:30:00"},
		{"临近右端的表达式", "0 23 * * *", 1, "2026-09-16 23:00:00"},
		{"每分钟一次应满 1440 次", "* * * * *", 1440, "2026-09-16 00:00:00"},
		{"窗口内无周一则为合法空集", "0 0 * * 1", 0, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := GetSchedulesBetween(c.spec, start, end)
			if err != nil {
				t.Fatalf("表达式 %q 解析失败: %v", c.spec, err)
			}
			if len(got) != c.wantLen {
				t.Fatalf("表达式 %q 命中 %d 次，期望 %d 次；实际: %v",
					c.spec, len(got), c.wantLen, formatTimes(got))
			}
			if c.wantLen == 0 {
				return
			}
			if head := got[0].Format("2006-01-02 15:04:05"); head != c.wantHead {
				t.Fatalf("表达式 %q 首个触发时间 = %s，期望 %s", c.spec, head, c.wantHead)
			}
		})
	}
}

func formatTimes(times []time.Time) []string {
	out := make([]string, 0, len(times))
	for _, tm := range times {
		out = append(out, tm.Format("2006-01-02 15:04:05"))
	}
	return out
}
