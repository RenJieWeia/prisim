package rules

import (
	"fmt"

	"github.com/renjie/prism-core/pkg/core/domain"
)

// RangeRule 实现数值范围检查
type RangeRule struct {
	Min    float64
	Max    float64
	Action domain.RuleAction
}

// Check 检查读数是否在范围内
// 返回值: (修正后的Reading, 是否通过valid, 错误信息)
func (r *RangeRule) Check(prev *domain.Reading, curr domain.Reading) (domain.Reading, bool, error) {
	if curr.Value >= r.Min && curr.Value <= r.Max {
		return curr, true, nil
	}

	// 触发规则: 超出范围
	switch r.Action {
	case domain.ActionCorrect:
		// 修正策略: 截断 (Clamp)
		// 如果你想做更复杂的修正(比如取均值)，可能需要更多的上下文或配置
		correctedReading := curr
		// 并没有修改原始 curr, 而是返回拷贝
		if curr.Value < r.Min {
			correctedReading.Value = r.Min
		} else {
			correctedReading.Value = r.Max
		}
		// 关键: 这里我们没有直接修改 Quality 字段，因为 Reading 是原始读数结构。
		// 在 Domain Layer 设计中，Quality 通常是 Output (StandardResult) 的属性，或者我们需要在这个阶段引入。
		// 假设目前 Reading 只是 Raw，我们暂时只修改值。
		// 下游转换时，如果发现值变了，或者有额外标记，再决定 Quality。
		// 但为了简单，我们假设调用者知道这是被修正过的。
		return correctedReading, true, nil // True means "keep it"

	case domain.ActionReject:
		fallthrough
	default:
		return curr, false, fmt.Errorf("value %.2f out of range [%.2f, %.2f]", curr.Value, r.Min, r.Max)
	}
}
