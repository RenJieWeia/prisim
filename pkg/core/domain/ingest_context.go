package domain

import "context"

// IngestStrategy 定义数据摄入策略和优先级
// 用于解决“后到数据”与“已有数据”的冲突 (Backfilling & Conflict Resolution)
type IngestStrategy string

const (
	// IngestStrategyRealtime 实时流数据 (默认)
	// 场景：设备实时上传
	// 优先级：中等 (100)
	IngestStrategyRealtime IngestStrategy = "REALTIME"

	// IngestStrategyBatchLate 延迟批处理/补传
	// 场景：设备离线后上线，批量上传历史数据
	// 优先级：中低 (50) -> 通常不覆盖实时数据，除非实时数据缺失
	IngestStrategyBatchLate IngestStrategy = "BATCH_LATE"

	// IngestStrategyCalibration 人工校准/修正
	// 场景：管理员手动导入修正数据，或者算法重新计算的高精度数据
	// 优先级：最高 (1000) -> 强制覆盖
	IngestStrategyCalibration IngestStrategy = "CALIBRATION"
)

// IngestContext 携带摄入时的上下文信息
type IngestContext struct {
	TraceID  string
	Strategy IngestStrategy
	Operator string // 操作人 (SYSTEM 或 具体User)
	BatchID  string // 批次号
}

// GetPriority 根据策略获取具体的优先级数值
// 数值越大，优先级越高 (Winner's Logic)
func (s IngestStrategy) GetPriority() int {
	switch s {
	case IngestStrategyCalibration:
		return 1000
	case IngestStrategyRealtime:
		return 100
	case IngestStrategyBatchLate:
		return 50
	default:
		return 0
	}
}

type ingestContextKey struct{}

// NewContext returns a new Context that carries the IngestContext value.
func NewContext(ctx context.Context, info IngestContext) context.Context {
	return context.WithValue(ctx, ingestContextKey{}, info)
}

// FromContext returns the IngestContext value stored in ctx, if any.
func FromContext(ctx context.Context) (IngestContext, bool) {
	info, ok := ctx.Value(ingestContextKey{}).(IngestContext)
	return info, ok
}
