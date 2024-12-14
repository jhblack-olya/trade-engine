/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pingcap/log"
	"github.com/segmentio/kafka-go"
	"github.com/jhblack-olya/trade-engine/conf"
	"github.com/jhblack-olya/trade-engine/models"
)

var productId2Writer sync.Map

func getWriter(productId string) *kafka.Writer {
	writer, found := productId2Writer.Load(productId)
	if found {
		return writer.(*kafka.Writer)
	}

	gbeConfig := conf.GetConfig()

	newWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      gbeConfig.Kafka.Brokers,
		Topic:        TopicOrderPrefix + productId,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 5 * time.Millisecond,
	})
	productId2Writer.Store(productId, newWriter)
	return newWriter
}

func (e *Engine) SubmitOrder(order *models.Order) {
	buf, err := json.Marshal(order)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = getWriter(e.productId).WriteMessages(context.Background(), kafka.Message{Value: buf})
	if err != nil {
		log.Error(err.Error())
		models.KafkaErrCh <- err

	}
}

func SubmitOrder(order *models.Order) {
	buf, err := json.Marshal(order)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = getWriter(order.ProductId).WriteMessages(context.Background(), kafka.Message{Value: buf})
	if err != nil {
		log.Error(err.Error())
		models.KafkaErrCh <- err

	}
}
