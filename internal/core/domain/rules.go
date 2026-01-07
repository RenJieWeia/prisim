package domain

import (
	"fmt"
)

// CleaningRule 清洗规则接口 (Strategy Pattern)
// 定义单一的校验逻辑
type CleaningRule interface {
	// Check 校验当前读数
	// prev: 上一个保留下来的有效读数 (可能为 nil)
	// curr: 当前待校验的读数
	// 返回: passed (是否通过), err (拒绝原因/报警信息)
	Check(prev *Reading, curr Reading) (bool, error)
}

// --- 具体规则实现 (Concrete Strategies) ---

// 1. MonotonicRule 单调性规则
// 用于检测由设备重置或故障导致的读数回退
type MonotonicRule struct{}

func (r *MonotonicRule) Check(prev *Reading, curr Reading) (bool, error) {
	// 基本校验：读数不能为负
	if curr.Value < 0 {
		return false, fmt.Errorf("negative value: %.2f", curr.Value)
	}
	// 上下文校验：不能小于上一次 (除非我们要处理翻转，这里简化为严格单调)
	if prev != nil && curr.Value < prev.Value {
		return false, fmt.Errorf("value regression: current %.2f < prev %.2f", curr.Value, prev.Value)
	}
	return true, nil
}

// 2. JumpRule 跳变规则
// 检测数据增长是否过快 (超出阈值)
type JumpRule struct {
	MaxThreshold float64
}

func (r *JumpRule) Check(prev *Reading, curr Reading) (bool, error) {
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
// 检测数据变化是否过小 (例如: 传感器卡死)
// 注意: 这个规则通常用于报警，但不一定丢弃数据。
// 在本清洗链中，如果用户配置了此规则，意味着他们认为微小变化是无效噪音或故障，需要过滤。
type StagnationRule struct {
	MinThreshold float64
}

func (r *StagnationRule) Check(prev *Reading, curr Reading) (bool, error) {
	if prev == nil {
		return true, nil
	}
	diff := curr.Value - prev.Value
	if diff < r.MinThreshold && diff >= 0 { // 假设单调性已检查，只看正向微小变化
		return false, fmt.Errorf("value stagnation: diff %.4f < min %.4f", diff, r.MinThreshold)
	}
	return true, nil
}
