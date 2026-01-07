package domain

import (
	"time"
)

// Aligner 定义时间对齐能力的接口
// 核心职责：从散乱的时间序列中提取特定时间点的快照
type Aligner interface {
	FindSnapshot(readings []Reading, target time.Time) *Reading
}

// TimeAligner 默认实现：基于时间容差的最近邻查找
type TimeAligner struct {
	Tolerance time.Duration
}

func NewAligner(tolerance time.Duration) Aligner {
	return &TimeAligner{Tolerance: tolerance}
}

// FindSnapshot 寻找最接近 target 时间点的读数
func (t *TimeAligner) FindSnapshot(readings []Reading, target time.Time) *Reading {
	var best *Reading
	var minDiff time.Duration = -1

	for i := range readings {
		r := &readings[i]
		diff := r.Timestamp.Sub(target)
		if diff < 0 {
			diff = -diff
		}

		if diff <= t.Tolerance {
			if best == nil || diff < minDiff {
				best = r
				minDiff = diff
			}
		}
	}
	return best
}

// Helper code usually accompanying Aligner in strategies.go was reading full content?
// Let's check if I missed anything from strategies.go.
// I see FindSnapshot logic.
// Is there more code in strategies.go? Let's check.
