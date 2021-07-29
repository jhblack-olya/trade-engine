package matching

import (
	"fmt"

	logger "github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/models"
)

type Engine struct {
	productId   string
	orderReader OrderReader
	orderOffset int64
	orderCh     chan *offsetOrder
	logCh       chan Log
	OrderBook   *orderBook
}

type offsetOrder struct {
	Offset int64
	Order  *models.Order
}

func NewEngine(product *models.Product, orderReader OrderReader) *Engine {
	e := &Engine{
		productId:   product.Id,
		orderReader: orderReader,
		orderCh:     make(chan *offsetOrder, 10000),
	}
	return e
}

func (e *Engine) Start() {
	go e.runFetcher()
}

//runFetcher: go routine responsible for continuously pulling order from kafka topic pushed from orderapi
//and pushing into order channel
func (e *Engine) runFetcher() {
	var offset = e.orderOffset
	if offset > 0 {
		offset += 1
	}
	err := e.orderReader.SetOffset(offset)
	if err != nil {
		logger.Fatalf("set order reader offset error: %v", err)
	}

	for {
		offset, order, err := e.orderReader.FetchOrder()
		if err != nil {
			logger.Error(err)
			continue
		}
		e.orderCh <- &offsetOrder{offset, order}
	}
}

func (e *Engine) runApplier() {
	var orderOffset int64

	for {
		select {
		case offsetOrder := <-e.orderCh:
			var logs []Log
			if offsetOrder.Order.Status == models.OrderStatusCancelling {
				fmt.Println("logs = e.OrderBook.CancelOrder(offsetOrder.Order)")
			} else {
				logs = e.OrderBook.ApplyOrder(offsetOrder.Order)
			}
			for _, log := range logs {
				e.logCh <- log
			}
			orderOffset = offsetOrder.Offset
			fmt.Println("orderOffset ", orderOffset)
		}
	}
}
