package main

import (
	"github.com/prometheus/common/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/pushing"
	"gitlab.com/gae4/trade-engine/rest"
	"gitlab.com/gae4/trade-engine/service"
	"gitlab.com/gae4/trade-engine/worker"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	gbeConfig := conf.GetConfig()
	go func() {
		log.Info(http.ListenAndServe("localhost:6000", nil))
	}()

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

	rest.StartServer()
	select {}
}
