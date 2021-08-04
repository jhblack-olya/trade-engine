package main

import (
	"github.com/prometheus/common/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/service"

	"net/http"
	_ "net/http/pprof"

	"gitlab.com/gae4/trade-engine/rest"
	"gitlab.com/gae4/trade-engine/worker"
)

func main() {
	gbeConfig := conf.GetConfig()

	go func() {
		log.Info(http.ListenAndServe("localhost:6000", nil))
	}()

	go models.NewBinLogStream().Start()

	matching.StartEngine()

	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		worker.NewTradeMaker(matching.NewKafkaLogReader("tradeMaker", product.Id, gbeConfig.Kafka.Brokers)).Start()
	}

	rest.StartServer()
	select {}
}
