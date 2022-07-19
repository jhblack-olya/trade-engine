/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"gitlab.com/gae4/trade-engine/models"
)

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

func (s *KafkaOrderReader) SetOffset(offset int64) error {
	err := s.orderReader.SetOffset(offset)
	if err != nil {
		models.KafkaErrCh <- err
		return err
	}
	return nil
}

func (s *KafkaOrderReader) FetchOrder() (offset int64, order *models.Order, err error) {
	message, err := s.orderReader.FetchMessage(context.Background())
	if err != nil {
		models.KafkaErrCh <- err
		return 0, nil, err
	}
	err = json.Unmarshal(message.Value, &order)
	if err != nil {
		return 0, nil, err
	}
	return message.Offset, order, nil
}
