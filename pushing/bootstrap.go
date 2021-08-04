package pushing

import (
	"fmt"
)

func StartServer() {
	fmt.Println("start server")
	// gbeConfig := conf.GetConfig()

	sub := newSubscription()

	newRedisStream(sub).Start()

	// fmt.Println("before products")
	// products, err := service.GetProducts()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("after products")

	// for _, product := range products {
	// 	fmt.Println("ranging", product.Id, gbeConfig.Kafka.Brokers)
	// 	newTickerStream(product.Id, sub, matching.NewKafkaLogReader("tickerStream", product.Id, gbeConfig.Kafka.Brokers)).Start()
	// }
}
