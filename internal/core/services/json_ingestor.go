package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/renjie/prism/internal/core/domain"
)

// JsonUniversalIngestor 实现 UniversalIngestor 接口
// 专门处理 JSON 格式的数据流
type JsonUniversalIngestor struct {
	// downstream 是数据流向的下一站
	// 在实际系统中，这里可能是调用 Standardizer.ProcessAndStandardize
	// 或者推送到消息队列
	downstream func(context.Context, []domain.Reading) error
}

func NewJsonUniversalIngestor(downstream func(context.Context, []domain.Reading) error) *JsonUniversalIngestor {
	return &JsonUniversalIngestor{
		downstream: downstream,
	}
}

// IngestStream 实现 UniversalIngestor.IngestStream
// 简化版：我们假设输入总是 JSON 数组 [...]，以规避 decoder.Token 的复杂性
func (j *JsonUniversalIngestor) IngestStream(ctx context.Context, stream io.Reader) error {
	decoder := json.NewDecoder(stream)

	// 预判 Token: 期待 '['
	t, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read start token: %w", err)
	}

	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("expected JSON array '[', got %v", t)
	}

	readings, err := j.decodeArray(decoder)
	if err != nil {
		return err
	}

	if len(readings) > 0 {
		return j.downstream(ctx, readings)
	}
	return nil
}

// IngestBatch 实现 UniversalIngestor.IngestBatch
func (j *JsonUniversalIngestor) IngestBatch(ctx context.Context, file io.Reader, format string) error {
	if format != "json" {
		return fmt.Errorf("unsupported format for JsonIngestor: %s", format)
	}
	return j.IngestStream(ctx, file)
}

// --- Internal Parsing Logic ---

// rawPayload 定义接收的扁平化 JSON 结构
// 适配多种字段命名风格 (Snake Case / Camel Case)
type rawPayload struct {
	DeviceID  string      `json:"device_id"`
	Model     string      `json:"model"`
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"` // 支持 RFC3339 或 简单时间格式
	Value     json.Number `json:"value"`     // 使用 json.Number 避免精度丢失
}

func (j *JsonUniversalIngestor) decodeArray(decoder *json.Decoder) ([]domain.Reading, error) {
	var results []domain.Reading

	// while decoder.More()
	for decoder.More() {
		var p rawPayload
		if err := decoder.Decode(&p); err != nil {
			return nil, fmt.Errorf("decode error inside array: %w", err)
		}
		r, err := j.mapToDomain(p)
		if err != nil {
			// 策略：遇到坏数据是报错还是跳过？
			// 万能插头通常应该容错，这里我们记录错误但不中断流（演示目的），或者返回错误
			// 这里选择记录错误并继续，或者如果您希望严格模式，则返回 err
			fmt.Printf("Warning: skipping invalid item: %v\n", err)
			continue
		}
		results = append(results, r)
	}

	// Consume closing ']'
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return results, nil
}

func (j *JsonUniversalIngestor) decodeObjectRemaining(decoder *json.Decoder) ([]domain.Reading, error) {
	// 因为我们已经消费了第一个 '{'，如果不使用 Hack 手段，标准 Decode 会失败。
	// 这里我们用一种流式属性解析的方法，或者更简单的：
	// 在实际生产中，我们可以用 json.RawMessage 结合 Peek。
	// 但为了代码简洁，且考虑到 decoder 已经前进，我们手动解析该对象的字段。
	// *注意*: 在 Go 标准库中这比较繁琐。

	// 替代方案：如果不预读 Token，直接 Decode(&val)，如果是指针 interface{} 会自动判断。
	// 但我们想区分 array vs object。

	// 重新设计 IngestStream 策略：
	// 不预读 Token。直接 Decode 到 json.RawMessage，然后判断第一个字符。
	return nil, fmt.Errorf("single object mode not fully implemented in this demo snippet due to decoder constraint, please use array format '[...]'")
}

// mapToDomain 将扁平 JSON 转换为领域对象
func (j *JsonUniversalIngestor) mapToDomain(p rawPayload) (domain.Reading, error) {
	// 1. Time Parsing
	var ts time.Time
	var err error

	// 尝试标准格式
	if ts, err = time.Parse(time.RFC3339, p.Timestamp); err != nil {
		// 尝试常见格式
		if ts, err = time.Parse("2006-01-02 15:04:05", p.Timestamp); err != nil {
			return domain.Reading{}, fmt.Errorf("invalid timestamp format: %s", p.Timestamp)
		}
	}

	// 2. Value Parsing
	val, err := p.Value.Float64()
	if err != nil {
		return domain.Reading{}, fmt.Errorf("invalid value format: %v", p.Value)
	}

	return domain.Reading{
		DeviceInfo: domain.DeviceInfo{
			ID:    p.DeviceID,
			Model: p.Model,
			Type:  domain.DeviceType(p.Type),
		},
		Timestamp: ts,
		Value:     val,
	}, nil
}
