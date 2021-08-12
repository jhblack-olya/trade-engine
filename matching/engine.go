package matching

import (
	"fmt"
	"time"

	logger "github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/models"
)

const duration = 1

type Engine struct {
	productId            string
	orderReader          OrderReader
	orderOffset          int64
	orderCh              chan *offsetOrder
	logCh                chan Log
	OrderBook            *orderBook
	logStore             LogStore
	snapshotStore        SnapshotStore
	snapshotReqCh        chan *Snapshot
	snapshotApproveReqCh chan *Snapshot
	snapshotCh           chan *Snapshot
	expiryCh             chan *offsetOrder
}

type Snapshot struct {
	OrderBookSnapshot orderBookSnapshot
	OrderOffset       int64
}

type offsetOrder struct {
	Offset int64
	Order  *models.Order
}

func NewEngine(product *models.Product, orderReader OrderReader, logStore LogStore, snapshotStore SnapshotStore) *Engine {
	e := &Engine{
		productId:            product.Id,
		OrderBook:            NewOrderBook(product),
		orderReader:          orderReader,
		orderCh:              make(chan *offsetOrder, 10000),
		logCh:                make(chan Log, 10000),
		snapshotStore:        snapshotStore,
		logStore:             logStore,
		snapshotReqCh:        make(chan *Snapshot, 32),
		snapshotApproveReqCh: make(chan *Snapshot, 32),
		snapshotCh:           make(chan *Snapshot, 32),
		expiryCh:             make(chan *offsetOrder, 10000),
	}

	snapshot, err := snapshotStore.GetLatest()
	if err != nil {
		logger.Fatalf("get latest snapshot error: %v", err)
	}
	if snapshot != nil {
		e.restore(snapshot)
	}

	return e
}

func (e *Engine) Start() {
	go e.runFetcher()
	go e.runApplier()
	go e.runCommitter()
	go e.runSnapshots()
	go e.countDownTimer()
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
		if order.Type == models.OrderTypeLimit && order.ExpiresIn > 0 {
			/*e.expiryMap[order.Id] = &offsetOrder{
				Offset: offset,
				Order:  order,
			}*/
			e.expiryCh <- &offsetOrder{
				Offset: offset,
				Order:  order,
			}

		}
		if order.Type == models.OrderTypeMarket {
			order.ExpiresIn = 0
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
				logs = e.OrderBook.CancelOrder(offsetOrder.Order)
			} else {
				logs = e.OrderBook.ApplyOrder(offsetOrder.Order)
			}
			for _, log := range logs {
				e.logCh <- log
			}
			orderOffset = offsetOrder.Offset

		case snapshot := <-e.snapshotReqCh:
			fmt.Println("delta= orderoffset - snapsjot.OrderOffset ", orderOffset, " - ", snapshot.OrderOffset, "= delta")
			delta := orderOffset - snapshot.OrderOffset
			if delta <= 3 {
				continue
			}
			logger.Infof("should take snapshot: %v %v-[%v]-%v->",
				e.productId, snapshot.OrderOffset, delta, orderOffset)
			snapshot.OrderBookSnapshot = e.OrderBook.Snapshot()
			snapshot.OrderOffset = orderOffset
			e.snapshotApproveReqCh <- snapshot
		}
	}
}

func (e *Engine) runCommitter() {
	var seq = e.OrderBook.logSeq
	var pending *Snapshot = nil
	var logs []interface{}
	for {
		select {
		case log := <-e.logCh:
			if log.GetSeq() <= seq {
				logger.Info("discard log seq=%v", seq)
				continue
			}
			seq = log.GetSeq()
			logs = append(logs, log)

			if len(e.logCh) > 0 && len(logs) < 100 {
				continue
			}

			err := e.logStore.Store(logs)
			if err != nil {
				panic(err)
			}
			logs = nil

			if pending != nil && seq >= pending.OrderBookSnapshot.LogSeq {
				e.snapshotCh <- pending
				pending = nil
			}

		case snapshot := <-e.snapshotApproveReqCh:
			fmt.Println("Seq ", seq, " snapshot.OrderBookSnapshot.LogSeq ", snapshot.OrderBookSnapshot.LogSeq)
			if seq >= snapshot.OrderBookSnapshot.LogSeq {
				fmt.Println("I am here becouse snapshot seq is less than or equal to seq")
				e.snapshotCh <- snapshot
				pending = nil
				continue
			}

			if pending != nil {
				logger.Info("discard snapshot request (seq=%v), new one (seq=%v) received", pending.OrderBookSnapshot.LogSeq, snapshot.OrderBookSnapshot.LogSeq)
			}
			pending = snapshot
		}
	}
}

func (e *Engine) runSnapshots() {
	// Order orderOffset at the last snapshot
	orderOffset := e.orderOffset
	fmt.Println("Called snapshot at order offset ", orderOffset)
	for {
		select {
		case <-time.After(30 * time.Second):
			// make a new snapshot request
			fmt.Println("making snapshot request after 30 second")
			e.snapshotReqCh <- &Snapshot{
				OrderOffset: orderOffset,
			}

		case snapshot := <-e.snapshotCh:
			// store snapshot
			fmt.Println("Storing snapshot")
			err := e.snapshotStore.Store(snapshot)
			if err != nil {
				logger.Warnf("store snapshot failed: %v", err)
				continue
			}
			logger.Infof("new snapshot stored :product=%v OrderOffset=%v LogSeq=%v",
				e.productId, snapshot.OrderOffset, snapshot.OrderBookSnapshot.LogSeq)

			// update offset for next snapshot request
			orderOffset = snapshot.OrderOffset
		}
	}
}

func (e *Engine) restore(snapshot *Snapshot) {
	logger.Infof("restoring: %+v", *snapshot)
	e.orderOffset = snapshot.OrderOffset
	e.OrderBook.Restore(&snapshot.OrderBookSnapshot)
}

/*
func (e *Engine) countDownTimer() {
	for {
		select {
		case <-time.After(time.Duration(duration) * time.Second):
			// After every 1 second decrement timer for limit order
			e.decrementer()

		}
	}

}

func (e *Engine) decrementer() {

	for key, val := range e.expiryMap {
		val.Order.ExpiresIn = val.Order.ExpiresIn - 1
		if val.Order.ExpiresIn == 0 {
			delete(e.expiryMap, key)
			val.Order.Status = models.OrderStatusCancelling
			val.Order.UpdatedAt = time.Now()
			SubmitOrder(val.Order)
			//e.orderCh <- &offsetOrder{val.Offset, val.Order, 1}
		} else {
			depth := e.OrderBook.depths[val.Order.Side]
			status := depth.UpdateDepth(key, val.Order.ExpiresIn)
			// if status false order not present in order book it may have completed or got cancelled prior
			if !status {
				delete(e.expiryMap, key)
			}
		}
	}
}

func (e *Engine) countDownTimer() {
	for {
		select {
		case <-time.After(time.Duration(duration) * time.Second):
			// After every 1 second decrement timer for limit order
			for key, val := range e.expiryMap {
				go func(key int64, val *offsetOrder) {
					select {
					case updated := <-e.decrementer1(val):
						if updated.Order.ExpiresIn == 0 {
							fmt.Println("Getting cancelled order ", key)
							delete(e.expiryMap, key)
							updated.Order.Status = models.OrderStatusCancelling
							updated.Order.UpdatedAt = time.Now()
							SubmitOrder(updated.Order)
							//e.orderCh <- &offsetOrder{val.Offset, val.Order, 1}
						} else {
							depth := e.OrderBook.depths[updated.Order.Side]
							status := depth.UpdateDepth(key, updated.Order.ExpiresIn)
							// if status false order not present in order book it may have completed or got cancelled prior
							if !status {
								delete(e.expiryMap, key)
							}
						}
					}

				}(key, val)

			}
		}
	}

}

func (e *Engine) decrementer1(val *offsetOrder) chan *offsetOrder {
	orderCh := make(chan *offsetOrder)
	val.Order.ExpiresIn -= 1
	orderCh <- val
	return orderCh
}*/

func (e *Engine) countDownTimer() {
	for {
		select {
		case o := <-e.expiryCh:
			go o.timed(e)
		}
	}

}
func (o *offsetOrder) timed(e *Engine) {

	flag := 0
	elapse := time.Duration(1) * time.Second
	//expiresIn := o.Order.ExpiresIn
	elapse1 := time.Duration(15) * time.Second
	for {
		select {
		case <-time.After(elapse):
			fmt.Println("============== deEAD END===================")
			fmt.Println("going to kill ", o.Order.Id, " its ", elapse)
			o.Order.Status = models.OrderStatusCancelling
			o.Order.UpdatedAt = time.Now()
			SubmitOrder(o.Order)
			flag = 1
		case <-time.After(elapse1):
			fmt.Println("********** 15 sec update***********")
			fmt.Println("life time of ", o.Order.Id, " is ", elapse, " remaining ", elapse-elapse1)
			o.Order.ExpiresIn = int64((time.Duration(o.Order.ExpiresIn) * time.Second) - elapse1)
			depth := e.OrderBook.depths[o.Order.Side]
			status := depth.UpdateDepth(o.Order.Id, o.Order.ExpiresIn)
			// if status false order not present in order book it may have completed or got cancelled prior
			if !status {
				flag = 1
			}
		}
		if flag == 1 {
			break
		}
	}
	/*for {
		select {
		case <-time.After(elapse):
			expiresIn -= 1
			if expiresIn == 0 {
				fmt.Println("============== deEAD END===================")
				fmt.Println("going to kill ", o.Order.Id, " its ", expiresIn)

				o.Order.Status = models.OrderStatusCancelling
				o.Order.UpdatedAt = time.Now()
				SubmitOrder(o.Order)
				flag = 1
			} else {
				depth := e.OrderBook.depths[o.Order.Side]
				status := depth.UpdateDepth(o.Order.Id, expiresIn)
				// if status false order not present in order book it may have completed or got cancelled prior
				if !status {
					flag = 1
				}
			}
		}
		if flag == 1 {
			break
		}
	}*/

}
