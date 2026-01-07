package rules

import (
	"fmt"

	"github.com/renjie/prism/internal/core/domain"
)

// --- Factory ---

func NewChain(configs []domain.RuleConfig) []domain.CleaningRule {
	var chain []domain.CleaningRule
	for _, cfg := range configs {
		switch cfg.ID {
		case "monotonic":
			chain = append(chain, &MonotonicRule{})
		case "jump":
			threshold, _ := getFloat(cfg.Params, "max_threshold")
			chain = append(chain, &JumpRule{MaxThreshold: threshold})
		case "stagnation":
			threshold, _ := getFloat(cfg.Params, "min_threshold")
			chain = append(chain, &StagnationRule{MinThreshold: threshold})
		}
	}
	return chain
}

func getFloat(params map[string]any, key string) (float64, bool) {
	if v, ok := params[key]; ok {
		// Handle JSON unmarshal float64/int/etc
		if f, ok := v.(float64); ok {
			return f, true
		}
		if i, ok := v.(int); ok {
			return float64(i), true
		}
	}
	return 0, false
}

// --- Concrete Strategies ---

// 1. MonotonicRule 单调性规则
type MonotonicRule struct{}

func (r *MonotonicRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
	if curr.Value < 0 {
		return false, fmt.Errorf("negative value: %.2f", curr.Value)
	}
	if prev != nil && curr.Value < prev.Value {
		return false, fmt.Errorf("value regression: current %.2f < prev %.2f", curr.Value, prev.Value)
	}
	return true, nil
}

// 2. JumpRule 跳变规则
type JumpRule struct {
	MaxThreshold float64
}

func (r *JumpRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
	if prev == nil {
		return true, nil
	}
	diff := curr.Value - prev.Value
	if diff > r.MaxThreshold {
		return false, fmt.Errorf("abnormal jump: diff %.2f > max %.2f", diff, r.MaxThreshold)
	}
	return true, nil
}

// 3. StagnationRule 停滞规则
type StagnationRule struct {
	MinThreshold float64
}

func (r *StagnationRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
	if prev == nil {
		return true, nil
	}
	diff := curr.Value - prev.Value
	if diff < r.MinThreshold && diff >= 0 {
		return false, fmt.Errorf("value stagnation: diff %.4f < min %.4f", diff, r.MinThreshold)
	}
	return true, nil
}
