package ports

import (
	"context"
	"time"

	"github.com/renjie/prism/internal/core/domain"
)

// StandardReadingRepository 标准读数仓储接口
// 对应核心竞争力: 输出“数据标准”的持久化载体
// 职责: 存储经过 Standardizer 清洗和对齐后的“黄金数据”，供下游查询整个园区/工厂的标准历史。
type StandardReadingRepository interface {
	// Save 保存单个标准读数
	Save(ctx context.Context, reading domain.StandardReading) error

	// SaveBatch 批量保存 (用于高吞吐流式写入)
	SaveBatch(ctx context.Context, readings []domain.StandardReading) error

	// FindExact 获取特定时间点的标准读数 (对应 GetStandardReading)
	// 场景: "获取 D1 设备在 10:00:00 的确切标准读数"
	FindExact(ctx context.Context, deviceID string, timestamp time.Time) (*domain.StandardReading, error)

	// FindRange 获取时间范围内的标准读数
	// 场景: 报表生成、趋势分析
	FindRange(ctx context.Context, deviceID string, start, end time.Time) ([]domain.StandardReading, error)
}

// DeviceRepository 设备元数据仓储 (可选，视校验需求而定)
type DeviceRepository interface {
	// Exists 检查设备是否存在
	Exists(ctx context.Context, deviceID string) (bool, error)
}
