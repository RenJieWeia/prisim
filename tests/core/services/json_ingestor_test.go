package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/renjie/prism/internal/core/domain"
	"github.com/renjie/prism/internal/core/services"
)

func TestJsonUniversalIngestor(t *testing.T) {
	// 1. Setup Data Source
	// Relocated to tests/core/services, relative path remains the same as internal/core/services for accessing root testdata
	// internal/core/services -> ../../../testdata
	// tests/core/services -> ../../../testdata
	file, err := os.Open("../../../testdata/stream_data.json")
	if err != nil {
		t.Fatalf("failed to open test data: %v", err)
	}
	defer file.Close()

	// 2. Setup "Downstream" Mock
	var received []domain.Reading
	mockDownstream := func(ctx context.Context, data []domain.Reading) error {
		received = append(received, data...)
		return nil
	}

	// 3. Initialize Ingestor
	ingestor := services.NewJsonUniversalIngestor(mockDownstream)

	// 4. Ingest
	err = ingestor.IngestStream(context.Background(), file)
	if err != nil {
		t.Fatalf("IngestStream failed: %v", err)
	}

	// 5. Verify
	if len(received) != 2 {
		t.Fatalf("expected 2 readings, got %d", len(received))
	}

	// Check Item 1 (RFC3339)
	r1 := received[0]
	if r1.DeviceInfo.ID != "D-Uni-01" {
		t.Errorf("Item 1 ID mismatch: got %s", r1.DeviceInfo.ID)
	}
	if r1.Value != 500.5 {
		t.Errorf("Item 1 Value mismatch: got %.2f", r1.Value)
	}

	// Check Item 2 (Simple Date)
	r2 := received[1]
	if r2.DeviceInfo.Type != "WATER" {
		t.Errorf("Item 2 Type mismatch: got %s", r2.DeviceInfo.Type)
	}
	if r2.Timestamp.Hour() != 12 || r2.Timestamp.Minute() != 30 {
		t.Errorf("Item 2 Timestamp parsed incorrectly: %v", r2.Timestamp)
	}
}
