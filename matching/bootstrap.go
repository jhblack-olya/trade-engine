package matching

import (
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/service"
)

func StartEngine() {
	gbeConfig := conf.GetConfig()

	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		orderReader := NewKafkaOrderReader(product.Id, gbeConfig.Kafka.Brokers)
		logStore := NewKafkaLogStore(product.Id, gbeConfig.Kafka.Brokers)
		matchEngine := NewEngine(product, orderReader, logStore)
		log.Info("orderReader", orderReader)
		log.Info("matchingEngine", matchEngine)
	}

	log.Info("match engine ok")
}
