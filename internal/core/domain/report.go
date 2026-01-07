package domain

import "time"

// ReportPeriod 报表统计维度
// 对应需求 2: 多维聚合 (小时、日、月)
type ReportPeriod string

const (
	ReportPeriodHour  ReportPeriod = "HOUR"
	ReportPeriodDay   ReportPeriod = "DAY"
	ReportPeriodMonth ReportPeriod = "MONTH"
)

// EnergyReport 能耗统计报表
// 对应需求 3.3: 步长聚合 & 分类统计
type EnergyReport struct {
	ID          string       `json:"id"`
	DeviceID    string       `json:"device_id"`
	DeviceModel string       `json:"device_model"`
	DeviceType  DeviceType   `json:"device_type"`
	Period      ReportPeriod `json:"period"`
	StartTime   time.Time    `json:"start_time"`  // 统计周期开始时间
	EndTime     time.Time    `json:"end_time"`    // 统计周期结束时间
	TotalUsage  float64      `json:"total_usage"` // 该周期内的总消耗

	// 对应需求: 统一度量衡（精度对齐）
	// 使用整型存储避免浮点数计算误差
	UsageScaled int64 `json:"usage_scaled"` // 缩放后的整数值 (e.g. 10.1234 -> 101234)
	ScaleFactor int   `json:"scale_factor"` // 缩放因子 (e.g. 10000)
}
