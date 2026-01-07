package domain

import "sort"

// Sanitizer 定义数据清洗能力的接口
// 核心职责：剔除脏数据（Duplicates, Outliers, Invalid Data）
type Sanitizer interface {
	Clean(readings []Reading) []Reading
}

// ChainSanitizer 默认实现：基于责任链模式的清洗器
// 它按顺序执行一系列 CleaningRule
type ChainSanitizer struct {
	rules []CleaningRule
}

// NewSanitizer 创建默认的基于规则链的清洗器
func NewSanitizer(rules ...CleaningRule) Sanitizer {
	return &ChainSanitizer{rules: rules}
}

// Clean 实现 Sanitizer 接口
func (s *ChainSanitizer) Clean(readings []Reading) []Reading {
	if len(readings) == 0 {
		return nil
	}

	// 1. 预处理：时间排序 (确保上下文校验的正确性)
	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	var clean []Reading
	var prev *Reading // 追踪上一个保留的有效值

	for _, curr := range readings {
		// 0. 内置规则: 时间戳去重 (完全相同时取第一个，或跳过后续)
		if prev != nil && prev.Timestamp.Equal(curr.Timestamp) {
			continue // Default Deduplication
		}

		// 执行规则链
		passed := true
		for _, rule := range s.rules {
			ok, _ := rule.Check(prev, curr)
			if !ok {
				passed = false
				// 可以在此处扩展日志记录
				break
			}
		}

		if passed {
			clean = append(clean, curr)
			prev = &clean[len(clean)-1]
		}
	}
	return clean
}
