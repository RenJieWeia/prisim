package services

import (
	"sort"

	"github.com/renjie/prism/internal/core/domain"
	"github.com/renjie/prism/internal/core/ports"
)

// ChainSanitizer 基于责任链模式的清洗器实现
type ChainSanitizer struct {
	rules []ports.CleaningRule
}

// NewSanitizer 创建默认的基于规则链的清洗器
func NewSanitizer(rules ...ports.CleaningRule) ports.Sanitizer {
	return &ChainSanitizer{rules: rules}
}

// Clean 实现 ports.Sanitizer 接口
func (s *ChainSanitizer) Clean(readings []domain.Reading) []domain.Reading {
	if len(readings) == 0 {
		return nil
	}

	// 1. 预处理：时间排序
	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	var clean []domain.Reading
	var prev *domain.Reading

	for _, curr := range readings {
		// 0. 内置规则: 时间戳去重
		if prev != nil && prev.Timestamp.Equal(curr.Timestamp) {
			continue // Default Deduplication
		}

		// 执行规则链
		passed := true
		for _, rule := range s.rules {
			ok, _ := rule.Check(prev, curr)
			if !ok {
				passed = false
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
