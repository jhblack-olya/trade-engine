/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package matching

import (
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/service"
)

var MatchEngine map[string]*Engine

func StartEngine() {
	gbeConfig := conf.GetConfig()

	products, err := service.GetProducts()
	if err != nil {
		panic(err)
	}
	MatchEngine = make(map[string]*Engine)
	for _, product := range products {
		orderReader := NewKafkaOrderReader(product.Id, gbeConfig.Kafka.Brokers)
		snapshotStore := NewRedisSnapshotStore(product.Id)
		logStore := NewKafkaLogStore(product.Id, gbeConfig.Kafka.Brokers)
		matchEngine := NewEngine(product, orderReader, logStore, snapshotStore)
		matchEngine.Start()
		MatchEngine[product.Id] = matchEngine

	}

	log.Info("match engine ok")
}
