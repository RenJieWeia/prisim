package ports

import "github.com/renjie/prism-core/pkg/core/domain"

// CleaningRule 清洗规则接口
// 这是一个策略接口，具体的业务规则（如单调性、跳变检测）由外部实现注入
type CleaningRule interface {
	Check(prev *domain.Reading, curr domain.Reading) (domain.Reading, bool, error)
}

// Sanitizer 数据清洗器接口
// 负责协调多个清洗规则的执行
type Sanitizer interface {
	// Clean 执行清洗逻辑，返回:
	// 1. clean: 通过规则的良品数据 (可能经过修正)
	// 2. quarantined: 违反规则被拒绝的次品数据 (包含拒绝原因)
	Clean(readings []domain.Reading) (clean []domain.Reading, quarantined []domain.QuarantineReading)
}
