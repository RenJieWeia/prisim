# Prism Core SDK

**Prism Core** 是 Prism 能源数据生态系统的基础 SDK。它提供了一个基于**六边形架构**的高性能、模块化数据处理引擎，专为水、电、气等异构能源数据的标准化而设计。

本库被设计为核心依赖 (Core Dependency)，供上层服务（如 HTTP API、CLI 工具、ETL 管道）引用，以提供一致的数据清洗和标准化能力。

## 🌟 核心特性

- **通用摄入 (Universal Ingestion)**: 
  - 支持基于流 (Stream) 的 JSON 解析，能够高效处理大规模数据集，内存占用极低。
- **稳健的数据清洗流水线 (Robust Pipeline)**:
  - **策略模式 (Strategy Pattern)**: 清洗规则完全解耦，支持热插拔。
  - **内置规则库**:
    - `MonotonicRule`: 单调性校验，防止累积读数出现负增长或异常回退。
    - `JumpRule`: 跳变检测，过滤掉物理上不可能的数值激增。
    - `StagnationRule`: 停滞检测，识别传感器死值或故障。
  - **责任链 (Chain of Responsibility)**: 通过 `Sanitizer` 串行执行配置的过滤器。
- **数据标准化 (Standardization)**:
  - **精度统一 (Unifier)**: 将浮点数转换为高精度的整型定点数 (Scaled Integer)，彻底消除浮点运算误差 (例如 kWh -> micro-kWh)。
  - **时间对齐 (Aligner)**: 将散乱的时间点对齐到标准的整点快照 (Snapshot)。
- **架构设计**:
  - **Domain (领域层)**: 纯粹的业务逻辑 (`pkg/core/domain`)，定义核心接口 (`CleaningRule`, `Sanitizer`, `Unifier`).
  - **Ports (端口层)**: 定义输入 (API/Ingestor) 和输出 (Repository) 的契约.
  - **Services (服务层)**: 编排领域逻辑与端口的胶水层 (`pkg/core/services`).

## 📂 项目结构

```
prism-core/
├── pkg/
│   └── core/
│       ├── domain/        # 核心业务逻辑 (实体 & 规则)
│       │   ├── aligner.go    # 时间对齐逻辑
│       │   ├── sanitizer.go  # 清洗器 (责任链)
│       │   ├── unifier.go    # 精度转换器
│       │   └── rules.go      # 具体清洗规则实现
│       ├── ports/         # 接口定义 (驱动/被驱动端口)
│       └── services/      # 应用服务 (流程编排)
├── tests/                 # 外部集成测试
│   ├── core/
│   │   ├── domain/        # 领域逻辑测试
│   │   └── services/      # 服务层测试
└── testdata/              # 测试用例样本数据
```

## 🚀 快速开始

### 安装

```bash
go get github.com/renjie/prism-core
```

使用 `JsonUniversalIngestor` 从文件或网络流中读取原始数据。

```go
import (
    "context"
    "os"
    "github.com/renjie/prism-core/pkg/core/services"
    "github.com/renjie/prism-core/pkg/core/domain"
)

// 定义数据接收回调（模拟“下游”处理）
downstreamHandler := func(ctx context.Context, readings []domain.Reading) error {
    for _, r := range readings {
        fmt.Printf("Received: %s - %.2f\n", r.DeviceInfo.ID, r.Value)
    }
    return nil
}

// 初始化摄入器
ingestor := services.NewJsonUniversalIngestor(downstreamHandler)

// 打开数据源 (io.Reader)
file, _ := os.Open("data.json")
defer file.Close()

// 开始流式处理
ingestor.IngestStream(context.Background(), file)
```

### 2. 配置清洗规则与标准化服务

核心服务 `CoreStandardizer` 负责编排清洗和标准化逻辑。你可以根据业务需求注入不同的规则。

```go
import (
    "github.com/renjie/prism-core/pkg/core/services"
    "github.com/renjie/prism-core/pkg/core/domain"
)

// 声明本地规则实现 (或从其他包导入)
type MyMonotonicRule struct{}
func (r *MyMonotonicRule) Check(prev *domain.Reading, curr domain.Reading) (bool, error) {
    if prev != nil && curr.Value < prev.Value {
        return false, fmt.Errorf("regression")
    }
    return true, nil
}

// 场景：我们需要严格的数据质量控制
// 使用 Functional Options 配置服务
standardizer := services.NewCoreStandardizer(
    services.WithPrecision(10000), 
    services.WithCleaningRules(&MyMonotonicRule{}),
    services.WithAlignment(15*time.Minute, 1*time.Minute),
)
``` 
)
```

### 3. 执行并获取结果

```go
rawReadings := []domain.Reading{
    {Timestamp: t1, Value: 100.0},
    {Timestamp: t2, Value: 90.0}, // 将被 MonotonicRule 过滤
    {Timestamp: t3, Value: 105.0},
}

// 执行标准化
results, err := standardizer.ProcessAndStandardize(ctx, rawReadings)

// 结果中只包含有效且转换后的数据
for _, res := range results {
    fmt.Printf("Standardized: %d (Raw: %.2f)\n", res.ValueScaled, res.ValueDisplay)
}
// Output:
// Standardized: 1000000 (Raw: 100.00)
// Standardized: 1050000 (Raw: 105.00)
```

### 4. 数据持久化 (Persistence)

项目提供了 SQLite 持久化适配器示例。

**前置条件**: 需要安装 CGO 支持的 SQLite 驱动 (推荐 GCC 环境)。
```bash
go get github.com/mattn/go-sqlite3
```

**示例代码**: 参见 `cmd/example/sqlite_demo/main.go`

```go
// 1. 初始化 SQLite 仓库
db, _ := sql.Open("sqlite3", "./prism.db")
repo, _ := sqlite.NewSqliteRepository(db)

// 2. 注入到 Standardizer
standardizer := services.NewCoreStandardizer(10000, repo, chain...)

// 3. 处理并自动保存
// ProcessAndStandardize 内部如果配置了 repo，会尝试保存结果
// (注: 当前 CoreStandardizer 实现可能需要更新以调用 Save，视具体实现而定，
//  默认 ProcessAndStandardize 主要是计算，持久化通常由应用层编排，
//  但在本示例架构中，CoreStandardizer 包含 repo 字段，可直接集成)
```

## 🛠 开发与测试

本项目采用严格的测试分离策略，单元测试位于 `tests/` 目录下。

### 运行测试
确保你已安装 Go 1.25+。

```bash
# 运行所有测试
go test ./tests/...

# 运行特定测试并查看详细输出
go test -v ./tests/core/services/
```

## 📄 许可证
MIT License
