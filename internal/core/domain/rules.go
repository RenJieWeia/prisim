package domain

// CleaningRule 清洗规则接口 (Strategy Pattern)
// 定义单一的校验逻辑
type CleaningRule interface {
	// Check 校验当前读数
	// prev: 上一个保留下来的有效读数 (可能为 nil)
	// curr: 当前待校验的读数
	// 返回: passed (是否通过), err (拒绝原因/报警信息)
	Check(prev *Reading, curr Reading) (bool, error)
}

// RuleConfig 规则配置项
// 用于从外部配置（JSON/YAML）动态构建清洗链
type RuleConfig struct {
	ID     string         `json:"id"`
	Params map[string]any `json:"params"`
}
