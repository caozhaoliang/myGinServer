package utils

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// ValidateCronExpr 校验标准 5 字段表达式
func ValidateCronExpr(expr string) error {
	// ParseStandard 内部使用默认的 Parser，支持 5 字段和描述符
	_, err := cron.ParseStandard(expr)
	if err != nil {
		// 可以在这里包装错误，返回更友好的提示
		return fmt.Errorf("无效的 cron 表达式: %w", err)
	}
	return nil
}

// GetSchedulesBetween 返回在 [start, end) 区间内，该 cron 表达式所有的触发时间点
func GetSchedulesBetween(cronSpec string, start, end time.Time) ([]time.Time, error) {
	// 1. 解析 cron 表达式
	schedule, err := cron.ParseStandard(cronSpec)
	if err != nil {
		return nil, fmt.Errorf("解析表达式失败: %w", err)
	}

	var times []time.Time
	current := start

	for {
		// 2. 计算下一个触发时间
		next := schedule.Next(current)

		// 3. 终止条件：没有下一次，或者已超出结束时间
		if next.IsZero() || !next.Before(end) {
			break
		}

		// 4. 收集并步进
		times = append(times, next)
		current = next
	}

	return times, nil
}
