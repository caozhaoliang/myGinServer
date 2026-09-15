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
	// 2. 起点回退 1 秒再截断到整秒：
	//    cron 的 Next 返回的是「严格大于」入参的时间点，若直接用 start 起步，
	//    恰好落在 start 上的触发（如 0 0 * * * 落在窗口左侧的午夜整点）会被永久跳过。
	//    回退 1 秒后 Next 就能把 start 本身返回，使区间语义与 [start, end) 一致。
	current := start.Truncate(time.Second).Add(-time.Second)

	for {
		// 3. 计算下一个触发时间
		next := schedule.Next(current)

		// 4. 终止条件：没有下一次，或者已到达/越过右端（右端为开区间）
		if next.IsZero() || !next.Before(end) {
			break
		}

		// 5. 兜底：回退起步可能引入早于 start 的触发点，此处剔除，保证左端确为闭区间
		if !next.Before(start) {
			times = append(times, next)
		}
		current = next
	}

	return times, nil
}
