package domain

// IngestionResult 导入结果统计
type IngestionResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Skipped int      `json:"skipped"` // 重复或其他原因跳过
	Errors  []string `json:"errors"`  // 具体的错误信息
}
