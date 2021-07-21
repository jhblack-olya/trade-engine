package matching

import "github.com/segmentio/kafka-go"

const (
	TopicOrderPrefix = "matching_order_"
)

type KafkaOrderReader struct {
	orderReader *kafka.Reader
}

func NewKafkaOrderReader(productId string, brokers []string) *KafkaOrderReader {
	s := &KafkaOrderReader{}

	s.orderReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     TopicOrderPrefix + productId,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	return s
}
