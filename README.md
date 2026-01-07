# Prism Core SDK

[中文文档 (Chinese Documentation)](README_CN.md)

**Prism Core** is the foundational SDK for the Prism energy data ecosystem. It provides a highly modular, hexagonally architected engine for standardizing energy and utilities data (Water, Electricity, Gas) from heterogeneous sources.

This library is designed to be imported by other services (HTTP APIs, CLI tools, ETL pipelines) to provide consistent data processing capabilities.

## 🌟 Core Features

- **Universal Ingestion**: Stream-based JSON ingestor capable of handling large datasets efficiently with minimal memory footprint.
- **Robust Cleaning Pipeline**:
  - **Strategy Pattern** based cleaning rules.
  - **Pluggable Rules**:
    - `MonotonicRule`: Prevents negative accumulation/regressions.
    - `JumpRule`: Detects and filters impossible spikes.
    - `StagnationRule`: Identifies dead sensors.
  - **Chain of Responsibility**: `Sanitizer` runs a configurable chain of filters.
- **Data Standardization**:
  - **Precision Control**: `Unifier` converts floating-point readings to high-precision integer scaled values (e.g., kWh to micro-kWh) to eliminate floating-point arithmetic errors.
  - **Time Alignment**: `Aligner` snaps readings to standard intervals (Snapshots).
- **Hexagonal Architecture**:
  - **Domain**: Pure business logic (`pkg/core/domain`), standard interfaces (`CleaningRule`, `Sanitizer`, `Unifier`).
  - **Ports**: Inbound (API/Ingestors) and Outbound (Repositories/Databases) definitions.
  - **Services**: Orchestration layer gluing domain logic to ports (`pkg/core/services`).

## 📂 Project Structure

```
prism-core/
├── pkg/
│   └── core/
│       ├── domain/        # Pure Business Logic (Entities & Rules)
│       │   ├── aligner.go
│       │   ├── sanitizer.go
│       │   ├── unifier.go
│       │   └── rules.go
│       ├── ports/         # Interface Definitions (Driver/Driven)
│       └── services/      # Application Services (Orchestration)
├── tests/                 # External Integration Tests
│   ├── core/
│   │   ├── domain/
│   │   └── services/
└── testdata/              # Sample data for tests
```

## 🚀 Getting Started

### Installation

```bash
go get github.com/renjie/prism-core
```

### Usage Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    // Import from the public package path
    "github.com/renjie/prism-core/pkg/core/services"
    "github.com/renjie/prism-core/pkg/core/domain"
)

func main() {
    // 1. Setup Ingestion
    ingestor := services.NewJsonUniversalIngestor(func(ctx context.Context, readings []domain.Reading) error {
        fmt.Printf("Received batch of %d readings\n", len(readings))
        return nil
    })
    
    // 2. Setup Standardization Service
    // Configure with 15-minute alignment and 4-decimal precision
    standardizer := services.NewCoreStandardizer(
        services.WithAlignment(15*time.Minute, 1*time.Minute),
        services.WithPrecision(10000),
    )
}
```

### Running Tests

```bash
go test ./tests/...
```

## 🛠 Architecture

### domain.Sanitizer
The `Standardizer` service cleans incoming raw data using a chain of injected rules.

```go
// Define custom rules implementing ports.CleaningRule
type MaxLimitRule struct{ limit float64 }
func (r *MaxLimitRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
    if curr.Value > r.limit { return false, fmt.Errorf("exceeded") }
    return true, nil
}

// Injected via functional options
svc := services.NewCoreStandardizer(services.WithCleaningRules(&MaxLimitRule{100}))
```

### domain.Unifier
Handles the conversion between "Human Readable" floats and "Machine Precise" integers.

```go
// 4 decimal places precision (x10000)
unifier := domain.NewUnifier(10000) 
scaled := unifier.ToScaled(100.00019) // Result: 1000002
```

## 📄 License
MIT
