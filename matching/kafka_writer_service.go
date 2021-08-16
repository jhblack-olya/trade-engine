package matching

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pingcap/log"
	"github.com/segmentio/kafka-go"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/models"
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

func SubmitOrder(order *models.Order) {
	buf, err := json.Marshal(order)
	if err != nil {
		log.Error(err.Error())
		return
	}

	fmt.Println("order.ProductId", order.ProductId)
	err = getWriter(order.ProductId).WriteMessages(context.Background(), kafka.Message{Value: buf})
	if err != nil {
		log.Error(err.Error())
	}
}