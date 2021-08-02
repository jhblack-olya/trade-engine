package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/service"
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
		Topic:        matching.TopicOrderPrefix + productId,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 5 * time.Millisecond,
	})
	productId2Writer.Store(productId, newWriter)
	return newWriter
}

func PlaceOder(ctx *gin.Context) {
	var req placeOrderRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}

	side := models.Side(req.Side)
	if len(side) == 0 {
		side = models.SideBuy
	}

	orderType := models.OrderType(req.Type)
	if len(orderType) == 0 {
		orderType = models.OrderTypeLimit
	}

	if len(req.ClientOid) > 0 {
		_, err = uuid.Parse(req.ClientOid)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, newMessageVo(fmt.Errorf("invalid client_oid: %v", err)))
			return
		}
	}

	size := decimal.NewFromFloat(req.Size)
	price := decimal.NewFromFloat(req.Price)
	funds := decimal.NewFromFloat(req.Funds)

	order, err := service.PlaceOrder(req.UserId, req.ClientOid, req.ProductId, orderType,
		side, size, price, funds, req.ExpiresIn)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return
	}

	submitOrder(order)

	ctx.JSON(http.StatusOK, order)
}

func submitOrder(order *models.Order) {
	buf, err := json.Marshal(order)
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Println("order.ProductId", order.ProductId)
	err = getWriter(order.ProductId).WriteMessages(context.Background(), kafka.Message{Value: buf})
	if err != nil {
		log.Error(err)
	}
}
