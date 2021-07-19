package matching

import (
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/service"
)

func StartEngine() {
	// gbeConfig := conf.GetConfig()

	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		log.Info("product", product)
	}

	log.Info("match engine ok")
}
