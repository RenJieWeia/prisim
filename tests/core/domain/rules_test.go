package domain_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/renjie/prism/internal/core/domain"
	"github.com/renjie/prism/internal/core/ports"
	"github.com/renjie/prism/internal/core/services"
)

// --- Mock/local Rule Implementations for Tests ---

type monotonicRule struct{}

func (r *monotonicRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
	if curr.Value < 0 {
		return false, fmt.Errorf("negative")
	}
	if prev != nil && curr.Value < prev.Value {
		return false, fmt.Errorf("regression")
	}
	return true, nil
}

type jumpRule struct {
	max float64
}

func (r *jumpRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
	if prev == nil {
		return true, nil
	}
	if (curr.Value - prev.Value) > r.max {
		return false, fmt.Errorf("jump")
	}
	return true, nil
}

// TestChainIntegration now defines rules locally, proving isolation
func TestChainIntegration(t *testing.T) {
	// Setup chain: Monotonic AND MaxJump(100)
	sanitizer := services.NewSanitizer(
		&monotonicRule{},
		&jumpRule{max: 100},
	)

	tBase := time.Now()
	data := []domain.Reading{
		{Timestamp: tBase, Value: 100},
		{Timestamp: tBase.Add(1 * time.Minute), Value: 150}, // OK (+50)
		{Timestamp: tBase.Add(2 * time.Minute), Value: 140}, // Fail Monotonic (-10)
		{Timestamp: tBase.Add(3 * time.Minute), Value: 300}, // Fail Jump (+150 from 150)
		{Timestamp: tBase.Add(4 * time.Minute), Value: 200}, // OK (+50 from 150)
	}

	clean := sanitizer.Clean(data)

	if len(clean) != 3 {
		t.Errorf("Expected 3 items, got %d", len(clean))
	}
}

// Ensure the Clean() method receives correct data types 
func TestSanitizerInterface(t *testing.T) {
	var _ ports.Sanitizer = services.NewSanitizer()
}
