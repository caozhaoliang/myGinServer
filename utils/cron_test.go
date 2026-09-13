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
