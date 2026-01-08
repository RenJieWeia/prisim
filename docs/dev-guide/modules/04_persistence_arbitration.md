# 04. 持久化与冲突仲裁 (Persistence & Arbitration) 开发手册 v2.0

## 1. 核心理念：Winner-Takes-All

在工业数据场景下，数据库不是简单的“日志记录器”，它是“单一事实来源 (Single Source of Truth)”。
由于数据可能通过多条路径到达（实时流、离线补录、人工修正），我们必须在数据库层面解决冲突。

**核心法则**: 只有 **高信任度 (High Priority)** 的数据才有资格覆盖 **低信任度** 的数据。

## 2. 冲突解决矩阵 (Decision Matrix)

| 场景 | 库中现有数据 (Old) | 新到达数据 (New) | 决策结果 | 解释 |
| :--- | :--- | :--- | :--- | :--- |
| **正常采集** | (空) | Realtime (100) | **INSERT** | 正常写入 |
| **重复采集** | Realtime (100) | Realtime (100) | **UPDATE** | 同级更新 (通常幂等) |
| **补录过期** | Calibration (1000) | BatchLate (50) | **IGNORE** | 补录的历史数据不可覆盖人工校准值 |
| **数据升级** | Estimated (0) | Realtime (100) | **UPDATE** | 实测值覆盖估算值 |
| **人工修正** | Realtime (100) | Calibration (1000) | **UPDATE** | 管理员具有最高解释权 |

## 3. SQL 实现参考 (PostgreSQL)

我们利用 PG 的 `ON CONFLICT` 和 `WHERE` 子句原子性地执行上述矩阵。

```sql
INSERT INTO standard_readings (
    device_id, timestamp, value_scaled, priority, ingested_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
ON CONFLICT (device_id, timestamp) 
DO UPDATE SET
    value_scaled = EXCLUDED.value_scaled,
    priority     = EXCLUDED.priority,
    ingested_at  = NOW()
WHERE 
    -- 核心保护逻辑: 只有新数据的Priority >= 旧数据时才执行 Update
    EXCLUDED.priority >= standard_readings.priority;
```

## 4. Repository 接口最佳实践

在 Go 代码中，我们强制要求调用者显式意识到他们在做什么。

```go
// ❌ 错误做法：隐藏了重要逻辑
// func Save(reading StandardReading) error 

// ✅ 正确做法：显式传递策略
func Save(ctx context.Context, reading domain.StandardReading, strategy ports.UpsertStrategy) error
```

### 策略选择指南
1.  **UpsertStrategyHighPriorityWins** (推荐):
    *   绝大多数业务场景。
    *   自动处理上述矩阵逻辑。
2.  **UpsertStrategyLastWriteWins** (慎用):
    *   仅用于初始化迁移，或者明确知道“我就是要强制覆盖一切”的特殊管理接口。

## 5. 常见坑点 (Pitfalls)

### 5.1 为什么我的 Update 返回 nil 但数据没变？
*   **原因**: 触发了 `WHERE priority` 拦截机制。新数据的优先级低于库中数据。
*   **调试**: 检查 `ingest_context` 是否正确传递？通过 SQL 查询该记录当前的 `priority` 值。

### 5.2 为什么 Priority 都是 0？
*   **原因**: 上游 Ingest 时忘记注入 Context，导致默认为 0。
*   **修复**: 参考 `01_ingestion_context.md` 确保 Context 传递。
