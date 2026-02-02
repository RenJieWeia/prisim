# Prism Core 开发手册索引

本目录包含了 Prism Core 核心系统的开发者指南。为了帮助新成员快速上手，手册按照核心领域模块进行了分类。我们的架构严格遵循**六边形架构 (Hexagonal Architecture)**，请在开发前先通过阅读以下模块文档熟悉各个核心组件的职责与设计意图。

## 最近更新 (v2.0)

- **接口优化**: `CleaningRule` 接口重构，使用 `CleaningContext` 和 `CheckResult` 结构体
- **性能优化**: `Aligner` 从 O(n) 线性扫描优化为 O(log n) 二分查找
- **精度转换**: 内置 `DefaultScaleFactor = 10000`，自动完成浮点到高精度整型的转换
- **并发安全**: 异步操作添加超时控制，错误使用 `errors.Join()` 聚合
- **可测试性**: `RuleFactory` 新增 `NewRuleFactory()` 构造函数支持测试注入

## 模块导航

### 1. [数据接入与上下文 (Ingestion & Context)](./modules/01_ingestion_context.md)
*   **核心类/接口**: `UniversalIngestor`, `IngestContext`, `IngestStrategy`
*   **适用场景**: 开发新的数据源接入适配器 (Adapters)，理解数据优先级的传递机制。

### 2. [数据标准化与核心服务 (Standardization Service)](./modules/02_standardization.md)
*   **核心类/接口**: `CoreStandardizer`, `Aligner`, `StandardReading`
*   **适用场景**: 理解浮点数到高精度整型的转换逻辑，修改时间对齐算法，理解主控逻辑。
*   **v2.0 更新**: 二分查找优化、内置精度转换、并发错误聚合

### 3. [数据清洗与治理 (Sanitizer & Quarantine)](./modules/03_sanitization_governance.md)
*   **核心类/接口**: `Sanitizer`, `CleaningRule`, `CleaningContext`, `CheckResult`, `RuleFactory`, `QuarantineReading`
*   **适用场景**: **开发新的数据清洗规则**，理解脏数据的去向，处理隔离区逻辑。
*   **v2.0 更新**: 新接口设计，结构化返回结果

### 4. [持久化与冲突仲裁 (Repository & Arbitration)](./modules/04_persistence_arbitration.md)
*   **核心类/接口**: `StandardReadingRepository`, `UpsertStrategy`
*   **适用场景**: 理解数据落库时的"防覆盖"保护机制，开发新的数据库适配器。

---

## 快速开发原则
1.  **依赖倒置**: 核心逻辑 (`pkg/core`) 不允许依赖适配器 (`pkg/adapters`)。
2.  **上下文优先**: 所有新增功能的入口必须考虑 `IngestContext`，确保操作人、优先级等元数据不丢失。
3.  **显式契约**: 所有的策略参数（如 `UpsertStrategy`）必须显式传递，拒绝隐式行为。
4.  **结构化结果**: 清洗规则必须返回 `CheckResult`，明确标识通过/拒绝/修正状态。
