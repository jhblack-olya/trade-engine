/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/models"
)

const (
	orderIdWindowCap = 10000
)

type evaluated struct {
	MakerOrderId int64
	TakerOrderId int64
	Price        decimal.Decimal
	EvaluatedAt  string
}
type orderBook struct {
	// one product corresponds to one order book
	product *models.Product

	// depths: asks & bids
	depths map[models.Side]*depth

	// strictly continuously increasing transaction ID, used for the primary key ID of trade
	tradeSeq int64

	// strictly continuously increasing log SEQ, used to write matching log
	logSeq int64

	// to prevent the order from being submitted to the order book repeatedly,
	// a sliding window de duplication strategy is adopted.
	orderIdWindow  Window
	DanglingOrders []*models.Order
	ArtTraded      map[int64]evaluated
	artDepths      map[int64]map[models.Side]*depth
}

type orderBookSnapshot struct {
	ProductId     string
	Orders        []BookOrder
	TradeSeq      int64
	LogSeq        int64
	OrderIdWindow Window
}

type depth struct {
	// all orders
	orders map[int64]*BookOrder

	// price first, time first order queue for order match
	// priceOrderIdKey -> orderId
	queue *treemap.Map
}

type priceOrderIdKey struct {
	price   decimal.Decimal
	orderId int64
}

type BookOrder struct {
	OrderId        int64
	Size           decimal.Decimal
	Funds          decimal.Decimal
	Price          decimal.Decimal
	Side           models.Side
	Type           int64
	ClientOid      string
	ExpiresIn      int64
	BackendOrderId string
	Art            int64
	UserId         int64
}

func (o *orderBook) nextLogSeq() int64 {
	o.logSeq++
	return o.logSeq
}

func (o *orderBook) nextTradeSeq() int64 {
	o.tradeSeq++
	return o.tradeSeq
}
func newBookOrder(order *models.Order) *BookOrder {
	return &BookOrder{
		OrderId:        order.Id,
		Size:           order.Size,
		Funds:          order.Funds,
		Price:          order.Price,
		Side:           order.Side,
		Type:           order.Type,
		ClientOid:      order.ClientOid,
		ExpiresIn:      order.ExpiresIn,
		BackendOrderId: order.BackendOrderId,
		Art:            order.Art,
		UserId:         order.UserId,
	}
}

func (d *depth) add(order BookOrder) {
	d.orders[order.OrderId] = &order
	d.queue.Put(&priceOrderIdKey{order.Price, order.OrderId}, order.OrderId)
	if models.Trigger != nil {
		models.Trigger <- order.Art
	}

}

func (d *depth) decrSize(orderId int64, size decimal.Decimal) error {
	order, found := d.orders[orderId]
	if !found {
		return fmt.Errorf("order %v not found on book", orderId)

	}

	if order.Size.LessThan(size) {
		return fmt.Errorf("order %v size %v less than %v", orderId, order.Size, size)
	}
	order.Size = order.Size.Sub(size)
	if order.Size.IsZero() {
		delete(d.orders, orderId)
		d.queue.Remove(&priceOrderIdKey{order.Price, order.OrderId})
	}

	if models.Trigger != nil {
		models.Trigger <- order.Art
	}
	return nil
}

//UpdateDepth: updates the order in orderbook depth
func (d *depth) UpdateDepth(orderId int64, timer int64) bool {
	order, found := d.orders[orderId]
	if !found {
		return false

	}
	order.ExpiresIn = timer

	return true
}

//ApplyOrder: Provides window of execution to a order and task related to matching is carried out
func (o *orderBook) ApplyOrder(order *models.Order) (logs []Log) {
	err := o.orderIdWindow.put(order.Id)
	if err != nil {
		log.Error(err)
		return logs
	}
	//taker who comes in with new order normally to do immediate buy or sell
	takerOrder := newBookOrder(order)
	//setting Market-Buy order to Infinite high and Market-sell order at zero.
	//which ensures that market prices will cross/execute
	if takerOrder.Type == models.OrderTypeMarket.Int() {
		if takerOrder.Side == models.SideBuy {
			takerOrder.Price = decimal.NewFromFloat(math.MaxFloat32)
		} else {
			takerOrder.Price = decimal.Zero
		}
	}
	var executedValue, filledSize, tackerActualSize /*, makerExecutedValue, makerFilledSize*/ decimal.Decimal
	var takermatchedAt string
	tackerActualSize = takerOrder.Size
	makerDepth := o.artDepths[takerOrder.Art][takerOrder.Side.Opposite()]
	for itr := makerDepth.queue.Iterator(); itr.Next(); {
		//maker who have already placed order normally not an immediate buyer or seller
		//ex trader who place limit order
		makerOrder := makerDepth.orders[itr.Value().(int64)]
		//if makerOrder.Art != takerOrder.Art {
		//	continue
		//}
		//check if buying price is greater than or equal to ask price
		//or
		//check if selling price is lesser than or equal to bid price
		// if any of them false break
		if (takerOrder.Side == models.SideBuy && takerOrder.Price.LessThan(makerOrder.Price)) ||
			(takerOrder.Side == models.SideSell && takerOrder.Price.GreaterThan(makerOrder.Price)) {
			break
		}

		//trade price
		var price = makerOrder.Price
		//trade size
		var size decimal.Decimal
		if takerOrder.Type == models.OrderTypeLimit.Int() ||
			(takerOrder.Type == models.OrderTypeMarket.Int() && takerOrder.Side == models.SideSell) {
			if takerOrder.Size.IsZero() {
				break
			}

			//take the minimum size of taker and maker as trade size
			size = decimal.Min(takerOrder.Size, makerOrder.Size)
			//adjust the size of taker order so that if there is no most available deal to complete taker order size then
			//remaining can be completed for next itteration
			takerOrder.Size = takerOrder.Size.Sub(size)
			executedValue = executedValue.Add(makerOrder.Price.Mul(size))
			filledSize = size.Add(filledSize)

		} else if takerOrder.Type == models.OrderTypeMarket.Int() && takerOrder.Side == models.SideBuy {
			if takerOrder.Funds.IsZero() {
				break
			}
			//takerActualSize = takerOrder.Size
			fmt.Println("taker actual size ", takerOrder.Size)
			//Understand it by example
			//Let marketprice = 5 and size=5 therefor fund of taker = 5x5=25
			// let most available price i.e maker price = 6 and size = 3
			// trade happens on the basis of makers price. If it is equal with market price it will execute on marketprice
			//So we divide funds on basis of maker price to know what size of trade will get executed at current maker price
			//takerSize=25/6 = 4 for ease of understanding
			takerSize := takerOrder.Funds.Div(price).Truncate(o.product.BaseScale)
			if takerSize.IsZero() {
				break
			}

			//taking minimum of takerSize and makerSize so trade gets completely filled
			//size=3
			size = decimal.Min(takerSize, makerOrder.Size)
			if takerOrder.Size.LessThan(size) {
				size = takerOrder.Size
			}
			fmt.Println("size evaluated ", size)
			//fund=3*6=18
			funds := size.Mul(price)
			//adjusting remaining fund for traker 25-18 = 7
			takerOrder.Funds = takerOrder.Funds.Sub(funds)
			//Here trade executed for 3 bid remaining 2 bid will be filled for next available maker
			// Now market price or latest trade price is 6

			executedValue = funds.Add(executedValue)
			filledSize = size.Add(filledSize)
			takerOrder.Size = takerOrder.Size.Sub(size)

		} else {
			log.Fatal("unknown orderType and side combination")
		}
		//adjust size of maker order or delete maker order if size is zero
		// according to above example for this itteration fetched maker order has been settled and will be
		//deleted from order book

		//makerExecutedValue = makerExecutedValue.Add(price.Mul(size))
		//makerFilledSize = size.Add(makerFilledSize)
		err := makerDepth.decrSize(makerOrder.OrderId, size)
		if err != nil {
			log.Fatal(err)
		}
		//orderPrice=
		// matched,write a log
		makermatchedAt := time.Now().Format("2006-01-02 15:04:05")
		takermatchedAt = makermatchedAt
		matchLog := newMatchLog(o.nextLogSeq(), o.product.Id, o.nextTradeSeq(), takerOrder, makerOrder, price, size, takerOrder.ExpiresIn, makerOrder.ExpiresIn, takerOrder.Art, makerOrder.Art, takermatchedAt, makermatchedAt)
		logs = append(logs, matchLog)
		o.ArtTraded[makerOrder.Art] = evaluated{Price: price, EvaluatedAt: makermatchedAt, MakerOrderId: makerOrder.OrderId, TakerOrderId: takerOrder.OrderId}
		log.Info("Last traded price ", o.ArtTraded)
		// maker is filled
		if makerOrder.Size.IsZero() {
			doneLog := newDoneLog(o.nextLogSeq(), o.product.Id, makerOrder, makerOrder.Size, models.DoneReasonFilled, makerOrder.ExpiresIn, makerOrder.Art, decimal.Zero, decimal.Zero, makermatchedAt)
			logs = append(logs, doneLog)
		} /*else {
			pendingLog := newPendingLog(o.nextLogSeq(), o.product.Id, makerOrder, makerOrder.Art)
			logs = append(logs, pendingLog)

		}*/
	}
	//If pogram controller break out of loop
	//check if taker is of type limit and commodity to be trade is greater than 0
	if takerOrder.Type == models.OrderTypeLimit.Int() && takerOrder.Size.GreaterThan(decimal.Zero) {
		//It may be possible that there is no cross happened for entire order or
		//there was only partial order cross.
		//so it will be added to order book and set next log sequence in order to execute this order in future
		//o.depths[takerOrder.Side].add(*takerOrder)
		o.artDepths[takerOrder.Art][takerOrder.Side].add(*takerOrder)
		//	models.Trigger = make(chan int64)
		//models.Trigger <- takerOrder.Art
		openLog := newOpenLog(o.nextLogSeq(), o.product.Id, takerOrder, takerOrder.ExpiresIn, takerOrder.Art)
		logs = append(logs, openLog)
	} else {
		//if marketorder and order dint execute cancel order
		//and if there is no more order to fullfill taker order and partial done

		//if takerorder is limit order and size<=0-->done
		//if takerorder is market order--->
		var remainingSize = takerOrder.Size
		var reason = models.DoneReasonFilled
		if takerOrder.Type == models.OrderTypeMarket.Int() {
			takerOrder.Price = decimal.Zero
			remainingSize = decimal.Zero
			if takerOrder.Side == models.SideSell && takerOrder.Size.GreaterThan(decimal.Zero) { /* ||
				(takerOrder.Side == models.SideBuy && takerOrder.Funds.GreaterThan(decimal.Zero))*/
				takermatchedAt = time.Now().Format("2006-01-02 15:04:05")
				reason = models.DoneReasonCancelled
				//newPendingLog(o.nextLogSeq(), o.product.Id, nil, remainingSize)
			} else if takerOrder.Side == models.SideBuy && takerOrder.Funds.GreaterThan(decimal.Zero) {
				fmt.Println("filled size is equal to ", filledSize)
				fmt.Println("takerorder size is equal to ", takerOrder.Size)
				if !takerOrder.Size.IsZero() && tackerActualSize.GreaterThan(takerOrder.Size) {
					takermatchedAt = time.Now().Format("2006-01-02 15:04:05")
					reason = models.DoneReasonPartial
				} else if tackerActualSize.Equal(takerOrder.Size) {
					takermatchedAt = time.Now().Format("2006-01-02 15:04:05")
					reason = models.DoneReasonCancelled
				}
			}
		}

		doneLog := newDoneLog(o.nextLogSeq(), o.product.Id, takerOrder, remainingSize, reason, takerOrder.ExpiresIn, takerOrder.Art, executedValue, filledSize, takermatchedAt)
		logs = append(logs, doneLog)
	}
	return logs
}

//CancelOrder: cancels the order and removes it from orderbook
func (o *orderBook) CancelOrder(order *models.Order) (logs []Log) {
	_ = o.orderIdWindow.put(order.Id)
	cancelledAt := time.Now().Format("2006-01-02 15:04:05")
	bookOrder, found := o.artDepths[order.Art][order.Side].orders[order.Id]
	if !found {
		return logs
	}

	// Order the size of all decr, equal to the remove operation
	remainingSize := bookOrder.Size
	err := o.artDepths[order.Art][order.Side].decrSize(order.Id, bookOrder.Size)
	if err != nil {
		panic(err)
	}
	var doneLog *DoneLog
	if remainingSize.GreaterThan(decimal.Zero) && order.Size.GreaterThan(remainingSize) {
		pendingLog := newPendingLog(o.nextLogSeq(), o.product.Id, order.Side, remainingSize, order.Id, order.Type, order.Art)
		logs = append(logs, pendingLog)
		doneLog = newDoneLog(o.nextLogSeq(), o.product.Id, bookOrder, remainingSize, models.DoneReasonPartial, order.ExpiresIn, order.Art, decimal.Zero, decimal.Zero, cancelledAt)
	} else if order.Size.Equal(remainingSize) {
		doneLog = newDoneLog(o.nextLogSeq(), o.product.Id, bookOrder, remainingSize, models.DoneReasonCancelled, order.ExpiresIn, order.Art, decimal.Zero, decimal.Zero, cancelledAt)
	}
	return append(logs, doneLog)
}

//Snapsot: Creates the snapshot
func (o *orderBook) Snapshot() orderBookSnapshot {
	lengthSell := 0
	lengthBuy := 0
	var sellOrderDepth map[int64]*BookOrder
	var buyOrderDepth map[int64]*BookOrder
	var sellBookOrder, buyBookOrder []*BookOrder
	var wg sync.WaitGroup
	for _, val := range o.artDepths {
		lengthSell += len(val[models.SideSell].orders)
		lengthBuy += len(val[models.SideBuy].orders)
		sellOrderDepth = val[models.SideSell].orders
		buyOrderDepth = val[models.SideBuy].orders
		wg.Add(2)
		go func() {
			defer wg.Done()
			for _, order := range sellOrderDepth {
				sellBookOrder = append(sellBookOrder, order)
			}
		}()
		go func() {
			defer wg.Done()
			for _, order := range buyOrderDepth {
				buyBookOrder = append(buyBookOrder, order)
			}
		}()
		wg.Wait()

	}
	snapshot := orderBookSnapshot{
		Orders:        make([]BookOrder, lengthSell+lengthBuy),
		LogSeq:        o.logSeq,
		TradeSeq:      o.tradeSeq,
		OrderIdWindow: o.orderIdWindow,
	}
	i := 0
	for _, order := range sellBookOrder {
		snapshot.Orders[i] = *order
		i++
	}

	for _, order := range buyBookOrder {
		snapshot.Orders[i] = *order
		i++
	}
	return snapshot
}

//Restore: restores orders from snapshot to order book
func (o *orderBook) Restore(snapshot *orderBookSnapshot) {
	o.logSeq = snapshot.LogSeq
	o.tradeSeq = snapshot.TradeSeq
	o.orderIdWindow = snapshot.OrderIdWindow
	if o.orderIdWindow.Cap == 0 {
		o.orderIdWindow = newWindow(0, orderIdWindowCap)
	}
	//creating object for snapshot orders during restoration
	for _, order := range snapshot.Orders {
		if _, ok := o.artDepths[order.Art]; !ok {
			o.artDepths[order.Art] = o.NewArtDepth()
			o.artDepths[order.Art][order.Side].add(order)
			danglingOrder := &models.Order{
				Id:        order.OrderId,
				ExpiresIn: order.ExpiresIn,
				Side:      order.Side,
				Size:      order.Size,
				Status:    models.OrderStatusOpen,
				Funds:     order.Funds,
				Type:      order.Type,
				ProductId: snapshot.ProductId,
				Art:       order.Art,
			}
			o.DanglingOrders = append(o.DanglingOrders, danglingOrder)
		} else {
			o.artDepths[order.Art][order.Side].add(order)
			danglingOrder := &models.Order{
				Id:        order.OrderId,
				ExpiresIn: order.ExpiresIn,
				Side:      order.Side,
				Size:      order.Size,
				Status:    models.OrderStatusOpen,
				Funds:     order.Funds,
				Type:      order.Type,
				ProductId: snapshot.ProductId,
				Art:       order.Art,
			}
			o.DanglingOrders = append(o.DanglingOrders, danglingOrder)
		}
	}
}

func priceOrderIdKeyAscComparator(a, b interface{}) int {
	aAsserted := a.(*priceOrderIdKey)
	bAsserted := b.(*priceOrderIdKey)

	x := aAsserted.price.Cmp(bAsserted.price)
	if x != 0 {
		return x
	}

	y := aAsserted.orderId - bAsserted.orderId
	if y == 0 {
		return 0
	} else if y > 0 {
		return 1
	} else {
		return -1
	}
}

func priceOrderIdKeyDescComparator(a, b interface{}) int {
	aAsserted := a.(*priceOrderIdKey)
	bAsserted := b.(*priceOrderIdKey)

	x := aAsserted.price.Cmp(bAsserted.price)
	if x != 0 {
		return -x
	}

	y := aAsserted.orderId - bAsserted.orderId
	if y == 0 {
		return 0
	} else if y > 0 {
		return 1
	} else {
		return -1
	}
}

//NewOrderBook: Initializes the orderbook
func NewOrderBook(product *models.Product) *orderBook {
	orderBook := &orderBook{
		product:       product,
		orderIdWindow: newWindow(0, orderIdWindowCap),
		ArtTraded:     make(map[int64]evaluated),
		artDepths:     make(map[int64]map[models.Side]*depth),
	}
	return orderBook
}

//NewArtDepth: creates orderbook depth for an art
func (o *orderBook) NewArtDepth() map[models.Side]*depth {
	asks := &depth{
		queue:  treemap.NewWith(priceOrderIdKeyAscComparator),
		orders: map[int64]*BookOrder{},
	}

	bids := &depth{
		queue:  treemap.NewWith(priceOrderIdKeyDescComparator),
		orders: map[int64]*BookOrder{},
	}
	depths := map[models.Side]*depth{models.SideBuy: bids, models.SideSell: asks}
	return depths
}
