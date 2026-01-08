package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/renjie/prism-core/pkg/adapters/factory"
	"github.com/renjie/prism-core/pkg/core/domain"
	"github.com/renjie/prism-core/pkg/core/ports"
)

// cleanWithDynamicRules 根据设备类型动态加载规则进行清洗
func (s *CoreStandardizer) cleanWithDynamicRules(ctx context.Context, readings []domain.Reading) ([]domain.Reading, []domain.QuarantineReading, error) {
	// 1. Group by DeviceType
	typeGroups := make(map[domain.DeviceType][]domain.Reading)
	for _, r := range readings {
		typeGroups[r.DeviceInfo.Type] = append(typeGroups[r.DeviceInfo.Type], r)
	}

	var result []domain.Reading
	var quarantined []domain.QuarantineReading
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(typeGroups))
	// 2. Process each type group concurrently (or sequentially, concurrency here is minor optimization)
	// Given we hit DB, concurrency is good.
	for dType, grp := range typeGroups {
		wg.Add(1)
		go func(dt domain.DeviceType, curReadings []domain.Reading) {
			defer wg.Done()

			// a. Load Rules
			domainRules, err := s.ruleRepo.ListEnabledByDeviceType(ctx, dt)
			if err != nil {
				errChan <- fmt.Errorf("load rules for %s failed: %w", dt, err)
				return
			}

			// b. Convert Rules
			var execRules []ports.CleaningRule
			ruleFactory := factory.GetRuleFactory()

			for _, dr := range domainRules {
				idx, err := ruleFactory.CreateRule(dr)
				if err != nil {
					// Log warning but continue? Or fail?
					// Strict mode: fail
					errChan <- fmt.Errorf("convert rule %s failed: %w", dr.ID, err)
					return
				}
				execRules = append(execRules, idx)
			}

			// c. Sanitize
			localSanitizer := NewSanitizer(execRules...)
			cleanedRows, rejectedRows := localSanitizer.Clean(curReadings)

			mu.Lock()
			result = append(result, cleanedRows...)
			quarantined = append(quarantined, rejectedRows...)
			mu.Unlock()
		}(dType, grp)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, nil, <-errChan
	}

	return result, quarantined, nil
}
