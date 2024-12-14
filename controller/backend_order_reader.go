/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package controller

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/jhblack-olya/trade-engine/models"
)

const (
	TopicBackendOrderPrefix = "backend_order"
)

type KafkaOrderReader struct {
	orderReader *kafka.Reader
}

func NewKafkaOrderReader(brokers []string) *KafkaOrderReader {
	s := &KafkaOrderReader{}

	s.orderReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     TopicBackendOrderPrefix,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	return s
}

func (s *KafkaOrderReader) SetOffset(offset int64) error {
	return s.orderReader.SetOffset(offset)
}

func (s *KafkaOrderReader) FetchOrder() (offset int64, order *models.PlaceOrderRequest, err error) {
	message, err := s.orderReader.FetchMessage(context.Background())
	if err != nil {
		return 0, nil, err
	}
	err = json.Unmarshal(message.Value, &order)
	if err != nil {
		return 0, nil, err
	}
	return message.Offset, order, nil
}

func (s *KafkaOrderReader) ReadLag() (int64, error) {
	lag, err := s.orderReader.ReadLag(context.Background())
	return lag, err
}
