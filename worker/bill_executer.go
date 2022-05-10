/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package worker

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/pingcap/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/service"
)

const fillWorkerNum = 10

type BillExecutor struct {
	workerChs [fillWorkerNum]chan *models.Bill
}

//NewBillExecuter executes operation for unsettled bills
func NewBillExecuter() *BillExecutor {
	f := &BillExecutor{
		workerChs: [fillWorkerNum]chan *models.Bill{},
	}

	for i := 0; i < fillWorkerNum; i++ {
		f.workerChs[i] = make(chan *models.Bill, 256)
		go func(idx int) {
			for {
				select {
				case bill := <-f.workerChs[idx]:
					err := service.ExecuteBill(bill.UserId, bill.Currency)
					if err != nil {
						log.Error(err.Error())
					}
				}
			}
		}(i)
	}
	return f
}

func (s *BillExecutor) Start() {
	go s.runMqListener()

}

func (s *BillExecutor) runMqListener() {
	gbeConfig := conf.GetConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     gbeConfig.Redis.Addr,
		Password: gbeConfig.Redis.Password,
		DB:       0,
	})

	for {
		ret := redisClient.BRPop(0, models.TopicBill)
		if ret.Err() != nil {
			log.Error(ret.Err().Error())
			continue
		}
		var bill models.Bill
		err := json.Unmarshal([]byte(ret.Val()[1]), &bill)
		if err != nil {
			panic(ret.Err())
		}
		s.workerChs[bill.UserId%fillWorkerNum] <- &bill
	}
}

func (s *BillExecutor) runInspector() {
	for {
		select {
		case <-time.After(1 * time.Second):
			bills, err := service.GetUnsettledBills()
			if err != nil {
				log.Error(err.Error())
				continue
			}
			for _, bill := range bills {
				s.workerChs[bill.UserId%fillWorkerNum] <- bill
			}
		}
	}
}
