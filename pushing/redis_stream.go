/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package pushing

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/siddontang/go-log/log"
	//"github.com/jhblack-olya/trade-engine/conf"
	//"github.com/jhblack-olya/trade-engine/models"
	//"github.com/jhblack-olya/trade-engine/utils"
   "github.com/jhblack-olya/trade-engine/models"
   "github.com/jhblack-olya/trade-engine/conf"
   "github.com/jhblack-olya/trade-engine/utils"
)

type redisStream struct {
	sub   *subscription
	mutex sync.Mutex
}

func newRedisStream(sub *subscription) *redisStream {
	return &redisStream{
		sub:   sub,
		mutex: sync.Mutex{},
	}
}

func (s *redisStream) Start() {
	fmt.Println("starting redis stream")
	gbeConfig := conf.GetConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     gbeConfig.Redis.Addr,
		Password: gbeConfig.Redis.Password,
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		models.RedisErrCh <- err
		panic(err)
	}

	go func() {
		for {
			ps := redisClient.Subscribe(models.TopicOrder)
			_, err := ps.Receive()
			if err != nil {
				log.Error(err)
				models.RedisErrCh <- err
				continue
			}

			for {
				select {
				case msg := <-ps.Channel():
					var order models.Order
					err := json.Unmarshal([]byte(msg.Payload), &order)
					if err != nil {
						continue
					}

					fmt.Println("publishing order:productId:userId", ChannelOrder.Format(order.ProductId, order.UserId))
					s.sub.publish(ChannelOrder.Format(order.ProductId, order.UserId), OrderMessage{
						UserId:        order.UserId,
						Type:          "order",
						Sequence:      0,
						Id:            utils.I64ToA(order.Id),
						Price:         order.Price.String(),
						Size:          order.Size.String(),
						Funds:         "0",
						ProductId:     order.ProductId,
						Side:          order.Side.String(),
						OrderType:     order.Type,
						CreatedAt:     order.CreatedAt.Format(time.RFC3339),
						FillFees:      order.FillFees.String(),
						FilledSize:    order.FilledSize.String(),
						ExecutedValue: order.ExecutedValue.String(),
						Status:        order.Status.String(),
						Settled:       order.Settled,
					})
				}
			}
		}
	}()

	go func() {
		for {
			ps := redisClient.Subscribe(models.TopicAccount)
			_, err := ps.Receive()
			if err != nil {
				log.Error(err)
				models.RedisErrCh <- err
				continue
			}

			for {
				select {
				case msg := <-ps.Channel():
					var account models.Account
					err := json.Unmarshal([]byte(msg.Payload), &account)
					if err != nil {
						continue
					}

					fmt.Println("publishing account", ChannelFunds.FormatWithUserId(account.UserId))

					s.sub.publish(ChannelFunds.FormatWithUserId(account.UserId), FundsMessage{
						Type:      "funds",
						Sequence:  0,
						UserId:    utils.I64ToA(account.UserId),
						Currency:  account.Currency,
						Hold:      account.Hold.String(),
						Available: account.Available.String(),
					})
				}
			}
		}
	}()
}
