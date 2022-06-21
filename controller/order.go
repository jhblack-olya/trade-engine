/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package controller

import (
	"log"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	logger "github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
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
	if offset > 0 {
		offset += 1
	}
	err := b.orderReader.SetOffset(offset)
	if err != nil {
		logger.Fatalf("set order reader offset error: %v", err)
	}

	for {
		_, order, err := b.orderReader.FetchOrder()
		if err != nil {
			continue
		}
		log.Printf("Fetched order \n %+v", order)
		b.PlaceOrder(order)
	}
}

func (b *BackendOrder) PlaceOrder(req *models.PlaceOrderRequest) {
	order := &models.Order{}
	var err error
	if req.Status != models.OrderStatusCancelling.String() {
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
		commission := decimal.NewFromFloat(req.Commission)
		commissionPercent := decimal.NewFromFloat(req.CommissionPercent)
		order, err = service.PlaceOrder(req.UserId, req.ClientOid, req.ProductId, orderType,
			side, size, price, funds, req.ExpiresIn, req.BackendOrderId, req.Art, commission, commissionPercent)

		if err != nil {
			return
		}
	} else {
		db := mysql.SharedStore()
		order, err = db.GetOrderById(req.OrderId)
		if err != nil {
			log.Println("get order error ", err.Error())
			return
		}
		if order.Status != models.OrderStatusOpen {
			return
		}
		order.Status = models.OrderStatusCancelling

	}
	matching.SubmitOrder(order)
}
