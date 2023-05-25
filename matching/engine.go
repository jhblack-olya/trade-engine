package matching

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/log"
	"github.com/shopspring/decimal"
	logger "github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/models"
)

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

// NewEngine Intitializes matching engine node
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

// runFetcher: go routine responsible for continuously pulling order from kafka topic pushed from orderapi
// and pushing into order channel
func (e *Engine) runFetcher() {
	var offset = e.orderOffset
	if offset > 0 {
		offset += 1
	}
	err := e.orderReader.SetOffset(offset)
	if err != nil {
		logger.Fatalf("set order reader offset error: %v", err)
	}

	//Sending snapshot orders to timed and applier before new order comes in
	if len(e.OrderBook.DanglingOrders) > 0 {
		for _, dOrder := range e.OrderBook.DanglingOrders {
			if dOrder.Type == models.OrderTypeLimit.Int() && dOrder.ExpiresIn > 0 {
				e.expiryCh <- &offsetOrder{offset, dOrder}

			}

		}
	}

	for {

		offset, order, err := e.orderReader.FetchOrder()
		if err != nil {
			logger.Error(err)
			continue
		}
		fmt.Println("order book art Depths ", e.OrderBook.artDepths)
		if len(e.OrderBook.artDepths) == 0 {
			e.OrderBook.artDepths = e.OrderBook.NewArtDepth()
		}
		if order.Type == models.OrderTypeLimit.Int() && order.ExpiresIn > 0 {
			e.expiryCh <- &offsetOrder{offset, order}
		} else if order.Type == models.OrderTypeLimit.Int() && order.ExpiresIn <= 0 && order.ExpiresIn != -1 {
			order.Status = models.OrderStatusCancelling // if limit order comes with expiry less than or equal to zero
			// cancel order if expiry is -1 its admin limit order which will not cancel
		}
		if order.Type == models.OrderTypeMarket.Int() {
			order.ExpiresIn = 0
		}
		e.orderCh <- &offsetOrder{offset, order}
	}
}

// runApplier: Manages order execution
func (e *Engine) runApplier() {
	var orderOffset int64

	for {
		select {
		case offsetOrder := <-e.orderCh:
			var logs []Log
			fmt.Println("order fetched ", offsetOrder.Order)

			if offsetOrder.Order.Status == models.OrderStatusCancelling {
				logs = e.OrderBook.CancelOrder(offsetOrder.Order)
			} else {
				fmt.Println("order getting passed to applyorder ", offsetOrder.Order)
				logs = e.OrderBook.ApplyOrder(offsetOrder.Order)
			}
			for _, log := range logs {
				e.logCh <- log
			}
			orderOffset = offsetOrder.Offset

		case snapshot := <-e.snapshotReqCh:
			delta := orderOffset - snapshot.OrderOffset
			if delta <= 1000 {
				continue
			}
			logger.Infof("should take snapshot: %v %v-[%v]-%v->",
				e.productId, snapshot.OrderOffset, delta, orderOffset)
			snapshot.OrderBookSnapshot = e.OrderBook.Snapshot()
			snapshot.OrderOffset = orderOffset
			snapshot.OrderBookSnapshot.ProductId = e.productId
			e.snapshotApproveReqCh <- snapshot
		}
	}
}

// runCommitter: generates sequence number for log, writes log data into kafka broker and approves snapshots
func (e *Engine) runCommitter() {
	var seq = e.OrderBook.logSeq
	var pending *Snapshot = nil
	var logs []interface{}
	for {
		select {
		case log := <-e.logCh:
			fmt.Println("log.GetSeq() ", log.GetSeq(), " <= ", "seq= ", seq)
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
			if seq >= snapshot.OrderBookSnapshot.LogSeq {
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

// countDownTimer: times the order
func (e *Engine) countDownTimer() {
	for {
		select {
		case o := <-e.expiryCh:
			//extract order and method with time
			depth := e.OrderBook.artDepths[o.Order.Side]
			go depth.timed(o, e)
		}
	}

}

// use depth
func (d *depth) timed(o *offsetOrder, e *Engine) {
	flag := 0
	elapse := time.Duration(1) * time.Second
	expiresIn := o.Order.ExpiresIn
	for {
		select {
		case <-time.After(elapse):
			expiresIn -= 1
			if expiresIn == 0 {
				o.Order.Status = models.OrderStatusCancelling
				o.Order.UpdatedAt = time.Now()
				o.Order.ExpiresIn = 0
				logger.Info("Order ", o.Order.Id, " expired")

				e.SubmitOrder(o.Order)
				flag = 1
				models.Trigger <- o.Order.ProductId

			} else {
				//	fmt.Println("order ", o.Order.Id, " expires in ", expiresIn)
				status := d.UpdateDepth(o.Order.Id, expiresIn)
				// if status false order not present in order book it may have completed or got cancelled prior
				if !status {
					flag = 1
				}
				//if tp, ok := d.orders[o.Order.Id]; ok {
				//	logger.Info("Order ", o.Order.Id, " checking expiry", tp.ExpiresIn)
				//}
			}
		}
		if flag == 1 {
			break
		}
	}

}

func (e *Engine) GetLimitOrders(side models.Side, art int64, size decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	var estimateAmt, mostAvailableAmt decimal.Decimal
	limitOrders := e.OrderBook.artDepths[side.Opposite()]
	if limitOrders == nil {
		log.Info("no orders available for art" + strconv.FormatInt(art, 10))
		return decimal.Zero, decimal.Zero, decimal.Zero
	}
	flag := 0
	sizeSum := decimal.Zero
	fmt.Println("\n\n I am in get limit order")
	fmt.Println("\n size sum", sizeSum)
	for itr := limitOrders.queue.Iterator(); itr.Next(); {
		orders := limitOrders.orders[itr.Value().(int64)]
		//	sizeSum = orders.Size.Add(sizeSum)
		//	fmt.Println("\n size sum", sizeSum)

		if flag == 0 {
			mostAvailableAmt = orders.Price
			flag = 1
		}
		if orders.Size.GreaterThanOrEqual(size) {
			estimateAmt = estimateAmt.Add(orders.Price.Mul(size))
			size = decimal.Zero
		} else {
			estimateAmt = estimateAmt.Add(orders.Price.Mul(orders.Size))
			size = size.Sub(orders.Size)
		}
		if size == decimal.Zero {
			break
		}
		//	fmt.Printf("\n orders %+v", orders)
	}
	for itr := limitOrders.queue.Iterator(); itr.Next(); {
		orders := limitOrders.orders[itr.Value().(int64)]
		sizeSum = orders.Size.Add(sizeSum)
		fmt.Println("\n size sum", sizeSum)
		fmt.Printf("\n orders %+v", orders)

	}
	return estimateAmt, mostAvailableAmt, sizeSum

}

func (e *Engine) LiveOrderBook() (map[string]decimal.Decimal, map[string]decimal.Decimal, decimal.Decimal) {
	var (
		bidMaxPrice decimal.Decimal
		askMinPrice decimal.Decimal
		usdSpace    decimal.Decimal
		askDepth    map[string]decimal.Decimal
		bidDepth    map[string]decimal.Decimal
	)

	flag := 0
	askOrders := e.OrderBook.artDepths[models.SideSell]
	bidOrders := e.OrderBook.artDepths[models.SideBuy]

	if askOrders != nil {
		fmt.Println("ask block")
		askDepth = make(map[string]decimal.Decimal)
		for itr := askOrders.queue.Iterator(); itr.Next(); {
			orders := askOrders.orders[itr.Value().(int64)]
			fmt.Println("price ", orders.Price, " size ", orders.Size)
			if flag == 0 {
				askMinPrice = orders.Price
				flag = 1
			}
			if val, ok := askDepth[orders.Price.String()]; ok {
				//delete(askDepth, orders.Price)
				askDepth[orders.Price.String()] = val.Add(orders.Size)
			} else {
				askDepth[orders.Price.String()] = orders.Size
			}
		}
	}

	flag1 := 0
	if bidOrders != nil {
		fmt.Println("bid block")

		bidSizeSum := decimal.Zero
		bidDepth = make(map[string]decimal.Decimal)

		for itr := bidOrders.queue.Iterator(); itr.Next(); {
			orders := bidOrders.orders[itr.Value().(int64)]
			fmt.Println("price ", orders.Price, " size ", orders.Size)
			bidSizeSum = orders.Size.Add(bidSizeSum)
			if flag1 == 0 {
				bidMaxPrice = orders.Price
				flag1 = 1
			}
			if val, ok := bidDepth[orders.Price.String()]; ok {
				//delete(bidDepth, orders.Price)
				bidDepth[orders.Price.String()] = val.Add(orders.Size)
			} else {
				bidDepth[orders.Price.String()] = orders.Size
			}
		}
	}
	if bidOrders == nil && askOrders != nil {
		fmt.Println("block 1")
		return askDepth, nil, askMinPrice
	} else if bidOrders != nil && askOrders == nil {
		fmt.Println("block 2")
		return nil, bidDepth, bidMaxPrice
	} else if bidOrders == nil && askOrders == nil {
		fmt.Println("block 3")
		return nil, nil, decimal.Zero
	}
	fmt.Println("bids ", bidDepth)
	fmt.Println("ask ", askDepth)
	usdSpace = askMinPrice.Sub(bidMaxPrice)
	return askDepth, bidDepth, usdSpace

}

// runSnapshots: stores snapshots
func (e *Engine) runSnapshots() {
	// Order orderOffset at the last snapshot
	orderOffset := e.orderOffset
	for {
		select {
		case <-time.After(30 * time.Second):
			// make a new snapshot request
			e.snapshotReqCh <- &Snapshot{
				OrderOffset: orderOffset,
			}

		case snapshot := <-e.snapshotCh:
			// store snapshot
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

// restore: restores orders from snapshot to order book
func (e *Engine) restore(snapshot *Snapshot) {
	logger.Infof("restoring: %+v", *snapshot)
	e.orderOffset = snapshot.OrderOffset
	e.OrderBook.Restore(&snapshot.OrderBookSnapshot)
}
