package adapters

import (
	"encoding/json"
	"fmt"

	goolog "v2.googo.io/goo-log"
)

// KafkaAdapter Kafka 适配器
// 注意：这是一个基础实现，实际使用时需要集成 kafka 客户端库
type KafkaAdapter struct {
	topic    string // Kafka topic
	producer any    // Kafka producer（需要根据实际使用的 kafka 库类型定义）
}

// KafkaConfig Kafka 适配器配置
type KafkaConfig struct {
	Topic    string // Kafka topic
	Producer any    // Kafka producer
}

// NewKafkaAdapter 创建 Kafka 适配器
// 这是一个示例实现，实际使用时需要根据具体的 kafka 客户端库进行调整
func NewKafkaAdapter(config KafkaConfig) *KafkaAdapter {
	return &KafkaAdapter{
		topic:    config.Topic,
		producer: config.Producer,
	}
}

// Write 写入日志到 Kafka
func (k *KafkaAdapter) Write(msg *goolog.Message) {
	// 构建消息
	message := k.buildMessage(msg)
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return
	}

	// 这里需要根据实际使用的 kafka 客户端库来实现发送逻辑
	// 例如使用 sarama、confluent-kafka-go 等
	_ = messageBytes
	_ = k.producer

	// 示例代码（需要根据实际库调整）:
	// if producer, ok := k.producer.(*sarama.SyncProducer); ok {
	//     producer.SendMessage(&sarama.ProducerMessage{
	//         Topic: k.topic,
	//         Value: sarama.ByteEncoder(messageBytes),
	//     })
	// }
}

// buildMessage 构建 Kafka 消息
func (k *KafkaAdapter) buildMessage(msg *goolog.Message) map[string]any {
	message := map[string]any{
		"timestamp": msg.Time.Format("2006-01-02 15:04:05"),
		"level":     goolog.LevelText[msg.Level],
		"message":   fmt.Sprint(msg.Message...),
	}

	// 添加标签
	if len(msg.Entry.Tags) > 0 {
		message["tags"] = msg.Entry.Tags
	}

	// 添加字段
	if len(msg.Entry.Data) > 0 {
		for _, field := range msg.Entry.Data {
			message[field.Field] = field.Value
		}
	}

	// 添加追踪信息
	if len(msg.Entry.Trace) > 0 {
		message["trace"] = msg.Entry.Trace
	}

	return message
}
