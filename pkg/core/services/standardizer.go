package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/renjie/prism-core/pkg/core/domain"
	"github.com/renjie/prism-core/pkg/core/ports"
)

// CoreStandardizer 核心数据标准化服务
// 实现了 EnergyDataStandardizer 接口
type CoreStandardizer struct {
	sanitizer        ports.Sanitizer
	unifier          domain.Unifier
	aligner          domain.Aligner
	standardInterval time.Duration
	concurrencyLimit int                             // 并发限制
	repo             ports.StandardReadingRepository // 可选持久层依赖
}

// StandardizerOption 定义配置选项函数 (Functional Option Pattern)
type StandardizerOption func(*CoreStandardizer)

// WithPrecision 设置精度因子 (默认 10000)
func WithPrecision(factor int) StandardizerOption {
	return func(s *CoreStandardizer) {
		s.unifier = domain.NewUnifier(factor)
	}
}

// WithAlignment 设置时间对齐参数 (默认 15m, 5m)
func WithAlignment(interval, tolerance time.Duration) StandardizerOption {
	return func(s *CoreStandardizer) {
		s.standardInterval = interval
		s.aligner = domain.NewAligner(tolerance)
	}
}

// WithRepository 设置持久层依赖
func WithRepository(repo ports.StandardReadingRepository) StandardizerOption {
	return func(s *CoreStandardizer) {
		s.repo = repo
	}
}

// WithCleaningRules 设置清洗规则
func WithCleaningRules(rules ...ports.CleaningRule) StandardizerOption {
	return func(s *CoreStandardizer) {
		s.sanitizer = NewSanitizer(rules...)
	}
}

// WithConcurrencyLimit 设置最大并发数 (默认 100)
func WithConcurrencyLimit(limit int) StandardizerOption {
	return func(s *CoreStandardizer) {
		if limit > 0 {
			s.concurrencyLimit = limit
		}
	}
}

// NewCoreStandardizer 初始化标准化服务
// 使用 Functional Options 模式进行配置
func NewCoreStandardizer(opts ...StandardizerOption) ports.EnergyDataStandardizer {
	// 默认配置
	s := &CoreStandardizer{
		sanitizer:        NewSanitizer(),                 // 默认无规则
		unifier:          domain.NewUnifier(10000),       // 默认精度 4 位小数
		aligner:          domain.NewAligner(time.Minute), // 默认容差 1m
		standardInterval: 15 * time.Minute,               // 默认间隔 15m
		concurrencyLimit: 100,                            // 默认并发 100
		repo:             nil,
	}

	// 应用选项
	for _, opt := range opts {
		opt(s)
	}

	return s
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

	// Step 3 (Optimization): Concurrency Strategy (Sharding by DeviceID)
	deviceGroups := make(map[string][]domain.Reading)
	for _, r := range cleanReadings {
		deviceGroups[r.DeviceInfo.ID] = append(deviceGroups[r.DeviceInfo.ID], r)
	}

	var standards []domain.StandardReading
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(deviceGroups))

	// Semaphore for bounded concurrency
	sem := make(chan struct{}, s.concurrencyLimit)

	for _, readings := range deviceGroups {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		go func(devReadings []domain.Reading) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			// Sort by timestamp ensures correct range determination
			sort.Slice(devReadings, func(i, j int) bool {
				return devReadings[i].Timestamp.Before(devReadings[j].Timestamp)
			})

			if len(devReadings) == 0 {
				return
			}

			// Step C: Frequency Alignment (Time Alignment)
			// Generate time grid based on standard interval
			startTime := devReadings[0].Timestamp.Truncate(s.standardInterval)
			endTime := devReadings[len(devReadings)-1].Timestamp
			// Align endTime to grid ceiling
			if rem := endTime.Sub(endTime.Truncate(s.standardInterval)); rem > 0 {
				endTime = endTime.Truncate(s.standardInterval).Add(s.standardInterval)
			} else {
				endTime = endTime.Truncate(s.standardInterval)
			}

			var groupStandards []domain.StandardReading

			for t := startTime; !t.After(endTime); t = t.Add(s.standardInterval) {
				// Context cancellation check (Fast fail)
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
				}

				// Find snapshot for this time slot
				snapshot := s.aligner.FindSnapshot(devReadings, t)
				if snapshot != nil {
					// Step 2: B. 单条转换
					sr := s.standardizeOne(*snapshot)
					sr.Timestamp = t // Force alignment to the grid time
					groupStandards = append(groupStandards, sr)
				}
			}

			mu.Lock()
			standards = append(standards, groupStandards...)
			mu.Unlock()

		}(readings)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan // Return first error
	}

	// Step 3: Persistence (if configured)
	if s.repo != nil && len(standards) > 0 {
		if err := s.repo.SaveBatch(ctx, standards); err != nil {
			return nil, fmt.Errorf("failed to persist standards: %w", err)
		}
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
