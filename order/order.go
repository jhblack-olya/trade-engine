package order

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	logger "github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/service"
)

type BackendOrder struct {
	orderReader OrderReader
	orderOffset int64
}

func ProcessOrder() {
	gbeConfig := conf.GetConfig()

	orderReader := NewKafkaOrderReader(gbeConfig.Kafka.Brokers)

	e := &BackendOrder{orderReader: orderReader}
	readLag, err := e.orderReader.ReadLag()
	if readLag > 0 {
		readLag = readLag - 1
	}
	fmt.Println("readLag--------------------------", readLag)
	if err != nil {
		logger.Fatalf("set read lag  error: %v", err)
	}
	e.orderOffset = readLag
	e.Start()
}

func (b *BackendOrder) Start() {
	go b.runFetcher()
}

func (b *BackendOrder) runFetcher() {
	var offset = b.orderOffset
	fmt.Println("b.orderOffset :: ", b.orderOffset)
	if offset > 0 {
		offset += 1
	}
	err := b.orderReader.SetOffset(offset)
	if err != nil {
		logger.Fatalf("set order reader offset error: %v", err)
	}

	for {
		offset, order, err := b.orderReader.FetchOrder()
		fmt.Println("offset, order ->", offset, order.BackendOrderId, err)
		if err != nil {
			continue
		}

		b.PlaceOrder(order)
	}
}

func (b *BackendOrder) PlaceOrder(req *models.PlaceOrderRequest) {

	side := models.Side(req.Side)
	if len(side) == 0 {
		side = models.SideBuy
	}

	orderType := models.OrderType(req.Type)
	if len(orderType) == 0 {
		orderType = models.OrderTypeLimit
	}

	if len(req.ClientOid) > 0 {
		_, err := uuid.Parse(req.ClientOid)
		if err != nil {
			return
		}
	}

	size := decimal.NewFromFloat(req.Size)
	price := decimal.NewFromFloat(req.Price)
	funds := decimal.NewFromFloat(req.Funds)

	order, err := service.PlaceOrder(req.UserId, req.ClientOid, req.ProductId, orderType,
		side, size, price, funds, req.ExpiresIn, req.BackendOrderId)

	if err != nil {
		return
	}

	matching.SubmitOrder(order)
}
