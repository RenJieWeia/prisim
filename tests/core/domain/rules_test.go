package domain_test

import (
	"log"
	"testing"
	"time"

	"github.com/renjie/prism/internal/core/domain"
)

func TestMonotonicRule(t *testing.T) {
	rule := &domain.MonotonicRule{}

	// Case 1: Increasing - Pass
	prev := &domain.Reading{Value: 100}
	curr := domain.Reading{Value: 110}
	if ok, _ := rule.Check(prev, curr); !ok {
		t.Errorf("Expected pass for increase, got fail")
	}

	// Case 2: Decrease - Fail
	curr.Value = 90
	if ok, _ := rule.Check(prev, curr); ok {
		t.Errorf("Expected fail for decrease, got pass")
	}

	// Case 3: Reset detected (if we implement logic for it, currently strict monotonic)
	// Current implementation: strict monotonic check, so it fails on any decrease.
}

func TestJumpRule(t *testing.T) {
	rule := &domain.JumpRule{MaxThreshold: 50}

	prev := &domain.Reading{Value: 100}

	// Case 1: Small jump - Pass
	curr := domain.Reading{Value: 130} // +30
	if ok, _ := rule.Check(prev, curr); !ok {
		t.Errorf("Expected pass for small jump, got fail")
	}

	// Case 2: Huge jump - Fail
	curr.Value = 200 // +100
	if ok, _ := rule.Check(prev, curr); ok {
		t.Errorf("Expected fail for huge jump, got pass")
	}

	// Case 3: Huge negative jump (if strictly checking absolute difference)
	// JumpRule Check: diff > MaxThreshold. curr - prev.
}

func TestStagnationRule(t *testing.T) {
	// Rule: Must change by at least 0.1
	rule := &domain.StagnationRule{MinThreshold: 0.1}

	prev := &domain.Reading{Value: 100}

	// Case 1: Change enough - Pass
	curr := domain.Reading{Value: 100.2}
	if ok, _ := rule.Check(prev, curr); !ok {
		t.Errorf("Expected pass for sufficient change, got fail")
	}

	// Case 2: Tiny change - Fail
	curr.Value = 100.05
	if ok, _ := rule.Check(prev, curr); ok {
		t.Errorf("Expected fail for tiny change (stagnation), got pass")
	}

	// Case 3: No change - Fail
	curr.Value = 100
	if ok, _ := rule.Check(prev, curr); ok {
		t.Errorf("Expected fail for no change, got pass")
	}
}

// TestChainIntegration integrates with DataSanitizer to verify the chain works
func TestChainIntegration(t *testing.T) {
	// Setup chain: Monotonic AND MaxJump(100)
	sanitizer := domain.NewSanitizer(
		&domain.MonotonicRule{},
		&domain.JumpRule{MaxThreshold: 100},
	)

	tBase := time.Now()
	data := []domain.Reading{
		{Timestamp: tBase, Value: 100},
		{Timestamp: tBase.Add(1 * time.Minute), Value: 150}, // OK (+50)
		{Timestamp: tBase.Add(2 * time.Minute), Value: 140}, // Fail Monotonic (-10)
		{Timestamp: tBase.Add(3 * time.Minute), Value: 300}, // Fail Jump (+150 from 150? no, from last valid... which was 150. So +150 > 100 -> Fail)
		{Timestamp: tBase.Add(4 * time.Minute), Value: 200}, // OK (+50 from 150)
	}

	clean := sanitizer.Clean(data)

	// Expected: 100, 150, 200
	if len(clean) != 3 {
		t.Errorf("Expected 3 items, got %d", len(clean))
		for i, r := range clean {
			log.Printf("[%d] %v", i, r.Value)
		}
	}

	if len(clean) >= 3 {
		if clean[2].Value != 200 {
			t.Errorf("Third item should be 200, got %v", clean[2].Value)
		}
	}
}
