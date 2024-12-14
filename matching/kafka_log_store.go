/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/jhblack-olya/trade-engine/models"
)

const (
	topicBookMessagePrefix = "matching_message_"
)

type KafkaLogStore struct {
	logWriter *kafka.Writer
}

func NewKafkaLogStore(productId string, brokers []string) *KafkaLogStore {
	s := &KafkaLogStore{}
	s.logWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topicBookMessagePrefix + productId,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 5 * time.Millisecond,
	})
	return s
}

func (s *KafkaLogStore) Store(logs []interface{}) error {
	var messages []kafka.Message
	for _, log := range logs {
		val, err := json.Marshal(log)
		if err != nil {
			return err
		}
		messages = append(messages, kafka.Message{Value: val})
	}
	err := s.logWriter.WriteMessages(context.Background(), messages...)
	if err != nil {
		models.KafkaErrCh <- err
		return err
	}
	return nil
}
