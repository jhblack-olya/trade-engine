/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package worker

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	lru "github.com/hashicorp/golang-lru"
	"github.com/siddontang/go/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/service"
)

type FillExecutor struct {
	// Used to receive the fill after sharding, and sharding according to orderId can reduce lock contention.
	workerChs [fillWorkerNum]chan *models.Fill
}

//NewFillExecutor executes operations for unsettled fills
func NewFillExecutor() *FillExecutor {
	f := &FillExecutor{
		workerChs: [fillWorkerNum]chan *models.Fill{},
	}

	for i := 0; i < fillWorkerNum; i++ {
		f.workerChs[i] = make(chan *models.Fill, 512)
		go func(idx int) {
			settledOrderCache, err := lru.New(1000)
			if err != nil {
				panic(err)
			}
			for {
				select {
				case fill := <-f.workerChs[idx]:
					if settledOrderCache.Contains(fill.OrderId) {
						continue
					}
					order, err := service.GetOrderById(fill.OrderId)
					if order == nil {
						log.Warnf("order not found: %v", fill.OrderId)
						continue
					}
					if order.Status == models.OrderStatusCancelled || order.Status == models.OrderStatusFilled {
						settledOrderCache.Add(order.Id, struct{}{})
						continue
					}
					err = service.ExecuteFill(fill.OrderId, fill.ExpiresIn, fill.Art)
					if err != nil {
						log.Error(err)
					}
				}
			}
		}(i)
	}
	return f
}

func (s *FillExecutor) Start() {
	go s.runInspector()
	go s.runMqListener()
}

// Listen for message queue notifications
func (s *FillExecutor) runMqListener() {
	gbeConfig := conf.GetConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     gbeConfig.Redis.Addr,
		Password: gbeConfig.Redis.Password,
		DB:       0,
	})

	for {
		ret := redisClient.BRPop(0, models.TopicFill)
		if ret.Err() != nil {
			log.Error(ret.Err())
			models.RedisErrCh <- ret.Err()
			continue
		} else {

			var fill models.Fill
			err := json.Unmarshal([]byte(ret.Val()[1]), &fill)
			if err != nil {
				log.Error(err)
				continue
			}

			// Modulate according to orderId for sharding, the same orderId will be assigned to a fixed chan
			s.workerChs[fill.OrderId%fillWorkerNum] <- &fill
		}
	}
}

// Polling the database regularly
func (s *FillExecutor) runInspector() {
	for {
		select {
		case <-time.After(1 * time.Second):
			fills, err := service.GetUnsettledFills(1000)
			if err != nil {
				log.Error(err)
				models.MysqlErrCh <- err
				continue
			}

			for _, fill := range fills {
				s.workerChs[fill.OrderId%fillWorkerNum] <- fill
			}
		}
	}
}
