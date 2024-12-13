/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package main

import (
	"sync"

	"github.com/prometheus/common/log"
	"gitlab.com/jhblack-olya/trade-engine/conf"
	"gitlab.com/jhblack-olya/trade-engine/controller"
	"gitlab.com/jhblack-olya/trade-engine/matching"
	"gitlab.com/jhblack-olya/trade-engine/models"
	"gitlab.com/jhblack-olya/trade-engine/pushing"
	"gitlab.com/jhblack-olya/trade-engine/rest"
	"gitlab.com/jhblack-olya/trade-engine/service"
	"gitlab.com/jhblack-olya/trade-engine/worker"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	gbeConfig := conf.GetConfig()
	go func() {
		log.Info(http.ListenAndServe("localhost:6000", nil))
	}()
	models.CommonError = make(map[string]string)
	models.RedisErrCh = make(chan error, 10)
	models.MysqlErrCh = make(chan error, 10)
	models.KafkaErrCh = make(chan error, 10)
	rest.ClientConn = make(map[string]map[string]*rest.WebsocketClient)
	models.Trigger = make(chan string, 4)
	models.Mu = new(sync.Mutex)
	go func() {
		for {
			select {
			case val := <-models.RedisErrCh:
				models.CommonError["redis"] = val.Error()
			case val := <-models.MysqlErrCh:
				models.CommonError["mysql"] = val.Error()
			case val := <-models.KafkaErrCh:
				models.CommonError["kafka"] = val.Error()
			}
		}
	}()

	controller.ProcessOrder()

	go models.NewBinLogStream().Start()

	matching.StartEngine()

	pushing.StartServer()

	//fillExecutor add partial filled order to bills termed as delay bill
	worker.NewFillExecutor().Start()
	//BillExecutor settles the unsettled bills
	worker.NewBillExecuter().Start()
	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		worker.NewFillMaker(matching.NewKafkaLogReader("fillMaker", product.Id, gbeConfig.Kafka.Brokers)).Start()
		worker.NewTradeMaker(matching.NewKafkaLogReader("tradeMaker", product.Id, gbeConfig.Kafka.Brokers)).Start()
	}
	go rest.Bridge()
	rest.StartServer()
	select {}
}
