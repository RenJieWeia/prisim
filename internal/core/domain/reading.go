package domain

import "time"

// ReadingType 定义读数类型
type ReadingType string

const (
	ReadingTypeRaw      ReadingType = "RAW"      // 原始读数
	ReadingTypeStandard ReadingType = "STANDARD" // 标准化读数
)

// QualityState 数据质量状态
type QualityState string

const (
	QualityValid        QualityState = "VALID"        // 有效
	QualityCorrected    QualityState = "CORRECTED"    // 已修正 (如去噪、插值)
	QualityEstimated    QualityState = "ESTIMATED"    // 估算值
	QualityInterpolated QualityState = "INTERPOLATED" // 插值生成 (频率对齐产物)
)

// Reading 代表一次原始读数
type Reading struct {
	DeviceInfo DeviceInfo `json:"device_info"`
	Timestamp  time.Time  `json:"timestamp"`
	Value      float64    `json:"value"` // 累积读数 (Cumulative Value)
}

// StandardReading 代表“数据标准”输出
// 对应核心竞争力: 帮下游平台“避坑” & “数据标准”
type StandardReading struct {
	DeviceID     string       `json:"device_id"`
	Timestamp    time.Time    `json:"timestamp"`     // 标准时间点 (e.g. 10:00:00)
	ValueScaled  int64        `json:"value_scaled"`  // 统一度量衡: 高精度整型值
	ScaleFactor  int          `json:"scale_factor"`  // 精度因子 (e.g. 10000)
	ValueDisplay float64      `json:"value_display"` // 展示用浮点值
	Quality      QualityState `json:"quality"`       // 数据质量标记
	SourceType   ReadingType  `json:"source_type"`   // 数据来源类型
}
