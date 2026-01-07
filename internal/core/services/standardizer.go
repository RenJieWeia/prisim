package services

import (
	"context"
	"fmt"
	"time"

	"github.com/renjie/prism/internal/core/domain"
	"github.com/renjie/prism/internal/core/ports"
)

// CoreStandardizer 核心数据标准化服务
// 实现了 EnergyDataStandardizer 接口
type CoreStandardizer struct {
	sanitizer ports.Sanitizer
	unifier   domain.Unifier
	repo      ports.StandardReadingRepository // 可选持久层依赖
}

// NewCoreStandardizer 初始化标准化服务
// 参数:
//   - precisionFactor: 精度因子 (B. 统一度量衡)
//   - repo: 持久层实现 (传入 nil 则仅支持 ProcessAndStandardize 纯计算模式)
//   - rules: 可变参数，传入需要应用的清洗规则
func NewCoreStandardizer(precisionFactor int, repo ports.StandardReadingRepository, rules ...ports.CleaningRule) ports.EnergyDataStandardizer {
	// 如果没有提供规则，可以使用默认规则集，或者留空
	// 这里我们选择直接传入
	return &CoreStandardizer{
		sanitizer: NewSanitizer(rules...),
		unifier:   domain.NewUnifier(precisionFactor),
		repo:      repo,
	}
}

// GetStandardReading 获取特定时间点的标准读数
// 职责：查询服务 (Query Service)
// 描述: “某设备在某时间点的标准读数是多少？” -> 清洗过、精度对齐的标准答案。
func (s *CoreStandardizer) GetStandardReading(ctx context.Context, deviceID string, timestamp time.Time) (*domain.StandardReading, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not configured: cannot query historical standards in stateless mode")
	}
	return s.repo.FindExact(ctx, deviceID, timestamp)
}

func (s *CoreStandardizer) ProcessAndStandardize(ctx context.Context, rawReadings []domain.Reading) ([]domain.StandardReading, error) {
	// Step 1: A. 数据清洗 (替别人做“脏活累活”)
	// 剔除空值、负值、重复值和异常跳变
	// 这一步是批量操作，因为清洗依赖上下文（如前后值的跳变）
	cleanReadings := s.sanitizer.Clean(rawReadings)

	var standards []domain.StandardReading

	for _, r := range cleanReadings {
		// Context cancellation check (Fast fail)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Step 2: B. 单条转换
		sr := s.standardizeOne(r)
		standards = append(standards, sr)
	}

	return standards, nil
}

// standardizeOne 封装单条数据的转换逻辑 (SR - Single Responsibility: Mapping)
func (s *CoreStandardizer) standardizeOne(r domain.Reading) domain.StandardReading {
	// 1. 精度对齐
	valScaled := s.unifier.ToScaled(r.Value)

	// 2. 结构封装
	return domain.StandardReading{
		DeviceID:     r.DeviceInfo.ID,
		Timestamp:    r.Timestamp,
		ValueScaled:  valScaled,
		ScaleFactor:  s.unifier.GetScaleFactor(),
		ValueDisplay: r.Value,
		SourceType:   domain.ReadingTypeStandard,
		Quality:      domain.QualityValid, // 经过清洗剩下的都是有效值
	}
}
