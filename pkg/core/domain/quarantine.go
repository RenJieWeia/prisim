package domain

import "time"

// QuarantineStatus 定义隔离记录的状态
type QuarantineStatus string

const (
	QuarantineStatusPending  QuarantineStatus = "PENDING"  // 待处理
	QuarantineStatusResolved QuarantineStatus = "RESOLVED" // 已解决 (已修正并重新入库)
	QuarantineStatusIgnored  QuarantineStatus = "IGNORED"  // 已忽略 (确认无效)
)

// QuarantineReading 代表一条被“隔离”审查的异常数据
// 当数据未通过 Sanitizer 清洗规则时，会被封装为此对象存入隔离区
type QuarantineReading struct {
	ID        string           `json:"id"`
	Reading   Reading          `json:"reading"`    // 原始读数快照
	Reason    string           `json:"reason"`     // 隔离原因 (e.g. "Value -50 below range min 0")
	RuleID    string           `json:"rule_id"`    // 触发的规则ID
	CreatedAt time.Time        `json:"created_at"` // 隔离时间
	UpdatedAt time.Time        `json:"updated_at"` // 更新时间
	Status    QuarantineStatus `json:"status"`     // 当前状态

	// 可选: 记录批次信息，方便批量重试
	BatchID string `json:"batch_id,omitempty"`
}
