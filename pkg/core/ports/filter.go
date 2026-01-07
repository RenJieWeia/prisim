package ports

import "github.com/renjie/prism-core/pkg/core/domain"

// CleaningRule 清洗规则接口
// 这是一个策略接口，具体的业务规则（如单调性、跳变检测）由外部实现注入
type CleaningRule interface {
	Check(prev *domain.Reading, curr domain.Reading) (bool, error)
}

// Sanitizer 数据清洗器接口
// 负责协调多个清洗规则的执行
type Sanitizer interface {
	Clean(readings []domain.Reading) []domain.Reading
}
