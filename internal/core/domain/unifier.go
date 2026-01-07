package domain

import "math"

// Unifier 定义度量衡统一能力的接口
// 核心职责：处理数值精度和单位转换
type Unifier interface {
	ToScaled(val float64) int64
	FromScaled(val int64) float64
	GetScaleFactor() int
}

// MetricUnifier 默认实现：基于乘数因子的定点数转换
type MetricUnifier struct {
	Factor int // e.g., 10000 for 4 decimal places
}

func NewUnifier(factor int) Unifier {
	return &MetricUnifier{Factor: factor}
}

func (u *MetricUnifier) ToScaled(val float64) int64 {
	return int64(math.Round(val * float64(u.Factor)))
}

func (u *MetricUnifier) FromScaled(val int64) float64 {
	return float64(val) / float64(u.Factor)
}

func (u *MetricUnifier) GetScaleFactor() int {
	return u.Factor
}
