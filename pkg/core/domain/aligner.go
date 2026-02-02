package domain

import (
	"sort"
	"time"
)

// TimeAligner 默认实现：基于时间容差的二分查找
type TimeAligner struct {
	Tolerance time.Duration
}

// NewAligner 创建时间对齐器实例
// 返回的实例实现了 ports.Aligner 接口
func NewAligner(tolerance time.Duration) *TimeAligner {
	return &TimeAligner{Tolerance: tolerance}
}

// FindSnapshot 使用二分查找寻找最接近 target 时间点的读数
// 时间复杂度: O(log n)，前提是 readings 已按时间排序
func (t *TimeAligner) FindSnapshot(readings []Reading, target time.Time) *Reading {
	if len(readings) == 0 {
		return nil
	}

	// 二分查找: 找到第一个 Timestamp >= target 的位置
	idx := sort.Search(len(readings), func(i int) bool {
		return !readings[i].Timestamp.Before(target)
	})

	// 检查 idx 和 idx-1，取时间差更小的那个
	var best *Reading
	var minDiff time.Duration = t.Tolerance + 1

	candidates := []int{idx - 1, idx}
	for _, i := range candidates {
		if i >= 0 && i < len(readings) {
			diff := absDuration(readings[i].Timestamp.Sub(target))
			if diff <= t.Tolerance && diff < minDiff {
				best = &readings[i]
				minDiff = diff
			}
		}
	}

	return best
}

// absDuration 返回 Duration 的绝对值
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
