/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
	"gitlab.com/gae4/trade-engine/service"
	"gitlab.com/gae4/trade-engine/standalone"
)

func PlaceOrderAPI(ctx *gin.Context) {
	var req placeOrderRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}

	order := &models.Order{}
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
			ctx.JSON(http.StatusBadRequest, err.Error())

		}
	} else {
		db := mysql.SharedStore()
		order, err = db.GetOrderById(req.OrderId)
		if err != nil {
			log.Println("get order error ", err.Error())
			ctx.JSON(http.StatusNotFound, err.Error())

			return
		}
		if order.Status != models.OrderStatusOpen {
			ctx.JSON(http.StatusForbidden, "not allowed")

			return
		}
		order.Status = models.OrderStatusCancelling

	}
	matching.SubmitOrder(order)
	ctx.JSON(http.StatusOK, order)
}

type KafkaLogStore struct {
	logWriter *kafka.Writer
}

func NewKafkaLogStore(brokers []string) *KafkaLogStore {
	s := &KafkaLogStore{}
	s.logWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        "backend_order",
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
	return s.logWriter.WriteMessages(context.Background(), messages...)
}

func BackendOrder(ctx *gin.Context) {
	var req placeOrderRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}
	gbeConfig := conf.GetConfig()

	logStore := NewKafkaLogStore(gbeConfig.Kafka.Brokers)
	var logs []interface{}
	logs = append(logs, req)
	err = logStore.Store(logs)
	if err != nil {
		models.KafkaErrCh <- err
		ctx.JSON(http.StatusInternalServerError, "Failed to place order")
		return
	}

	ctx.JSON(http.StatusOK, "Order placed")
}

func EstimateAmount(ctx *gin.Context) {
	productId := ctx.Query("product_id")
	size, err := decimal.NewFromString(ctx.Query("size"))
	art, err := strconv.ParseInt(ctx.Query("art_name"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}
	side := ctx.Query("side")
	if err != nil || productId == "" || art == 0 || side == "" {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}
	estAmt, minAmt, sizeSum := standalone.GetEstimate(productId, size, art, models.Side(side))
	resp := estimateResponse{
		Amount:           estAmt,
		MostAvailableAmt: minAmt,
		DepthSize:        sizeSum,
	}
	ctx.JSON(http.StatusOK, resp)
}

var UpGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketClient struct {
	Ws        *websocket.Conn
	CloseChan chan bool
}

var ClientConn map[int64]map[int64]*WebsocketClient

func GetLiveOrderBook(ctx *gin.Context) {
	art, err := strconv.ParseInt(ctx.Query("art"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusForbidden, newMessageVo(err))

	}
	userId, err := strconv.ParseInt(ctx.Query("user"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusForbidden, newMessageVo(err))

	}
	product := ctx.Query("product")
	if product == "" {
		product = "ABT-USDT"
	}
	status := ctx.Query("status")
	ws, err := UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("error get connection")
		log.Fatal(err)
	}
	if status == "open" {
		//userClient[userId] = wsClient
		if userConn, ok := ClientConn[art]; ok {
			if _, ok1 := userConn[userId]; !ok1 {
				userConn[userId] = &WebsocketClient{
					Ws:        ws,
					CloseChan: make(chan bool),
				}
			}
		} else {
			ClientConn[art] = make(map[int64]*WebsocketClient)
			userClient := make(map[int64]*WebsocketClient)
			userClient[userId] = &WebsocketClient{
				Ws:        ws,
				CloseChan: make(chan bool),
			}
			ClientConn[art] = userClient
		}
		models.Trigger = make(chan int64, 1)
		models.Trigger <- art

	}
	go func() {
		for {
			select {
			case val := <-models.Trigger:
				if val > 0 {
					ask, bid, usdSpace := standalone.GetOrderBook(product, val)
					resp := models.OrderBookResponse{}
					resp.UsdSpace = usdSpace
					for key, val := range ask {
						mp := make(map[string]decimal.Decimal)
						mp[key] = val
						resp.Ask = append(resp.Ask, mp)
					}
					for key, val := range bid {
						mp := make(map[string]decimal.Decimal)
						mp[key] = val
						resp.Bid = append(resp.Bid, mp)
					}
					if conn, ok := ClientConn[val]; ok {
						for _, userConn := range conn {
							err := userConn.Ws.WriteJSON(&resp)
							if err != nil {
								log.Println("error write json: " + err.Error())
							}
						}

					}
				} else {
					break
				}
			case <-ClientConn[art][userId].CloseChan:
				if conn, ok := ClientConn[art]; ok {
					if userConn, ok := conn[userId]; ok {
						userConn.Ws.Close()
						delete(ClientConn[art], userId)
						close(models.UserChan[userId])
						delete(models.UserChan, userId)
						break
					}
				}
			}
		}
	}()

}

func CloseWebsocket(ctx *gin.Context) {
	art, err := strconv.ParseInt(ctx.Query("art"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusForbidden, newMessageVo(err))

	}
	userId, err := strconv.ParseInt(ctx.Query("user"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusForbidden, newMessageVo(err))

	}
	if conn, ok := ClientConn[art]; ok {
		if userConn, ok := conn[userId]; ok {
			userConn.CloseChan <- true
		}
	}
}
