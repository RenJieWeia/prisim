# Prism: Universal Energy Data Adapter

[中文文档 (Chinese Documentation)](README_CN.md)

**Prism** is a high-performance, modular data processing engine designed to standardize energy and utilities data (Water, Electricity, Gas) from heterogeneous sources. Built with **Hexagonal Architecture** (Ports & Adapters) principles in Go, strictly separating core domain logic from external dependencies.

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
  - **Domain**: Pure business logic, standard interfaces (`CleaningRule`, `Sanitizer`, `Unifier`).
  - **Ports**: Inbound (API/Ingestors) and Outbound (Repositories/Databases) definitions.
  - **Services**: Orchestration layer gluing domain logic to ports.

## 📂 Project Structure

```
prisim/
├── internal/
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

### Prerequisites
- Go 1.25+

### Running Tests
The project maintains a strict separation of unit/integration tests in the `tests/` directory.

```bash
go test ./tests/...
```

## 🛠 Architecture

### domain.Sanitizer
The `Standardizer` service cleans incoming raw data using a chain of injected rules.

```go
// Example Configuration
sanitizer := domain.NewSanitizer(
    &domain.MonotonicRule{}, 
    &domain.JumpRule{MaxThreshold: 100},
)
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
