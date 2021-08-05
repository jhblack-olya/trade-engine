package matching

import (
	"fmt"
	"math"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/models"
)

const (
	orderIdWindowCap = 10000
)

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
	orderIdWindow Window
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
	OrderId   int64
	Size      decimal.Decimal
	Funds     decimal.Decimal
	Price     decimal.Decimal
	Side      models.Side
	Type      models.OrderType
	ClientOid string
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
		OrderId:   order.Id,
		Size:      order.Size,
		Funds:     order.Funds,
		Price:     order.Price,
		Side:      order.Side,
		Type:      order.Type,
		ClientOid: order.ClientOid,
	}
}

func (d *depth) add(order BookOrder) {
	d.orders[order.OrderId] = &order
	d.queue.Put(&priceOrderIdKey{order.Price, order.OrderId}, order.OrderId)
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
	return nil
}

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
	if takerOrder.Type == models.OrderTypeMarket {
		if takerOrder.Side == models.SideBuy {
			takerOrder.Price = decimal.NewFromFloat(math.MaxFloat32)
		} else {
			takerOrder.Price = decimal.Zero
		}
	}
	//if taker are seller then makerDepth will be bids placed in order book and
	//if taker are buyer then makerDepth will be asks placed in order book
	makerDepth := o.depths[takerOrder.Side.Opposite()]

	for itr := makerDepth.queue.Iterator(); itr.Next(); {
		//maker who have already placed order normally not an immediate buyer or seller
		//ex trader who place limit order
		makerOrder := makerDepth.orders[itr.Value().(int64)]
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

		if takerOrder.Type == models.OrderTypeLimit ||
			(takerOrder.Type == models.OrderTypeMarket && takerOrder.Side == models.SideSell) {
			if takerOrder.Size.IsZero() {
				break
			}

			//take the minimum size of taker and maker as trade size
			size = decimal.Min(takerOrder.Size, makerOrder.Size)
			//adjust the size of taker order so that if there is no most available deal to complete taker order size then
			//remaining can be completed for next itteration
			takerOrder.Size = takerOrder.Size.Sub(size)

		} else if takerOrder.Type == models.OrderTypeMarket && takerOrder.Side == models.SideBuy {
			if takerOrder.Funds.IsZero() {
				break
			}
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
			//fund=3*6=18
			funds := size.Mul(price)
			//adjusting remaining fund for traker 25-18 = 7
			takerOrder.Funds = takerOrder.Funds.Sub(funds)
			//Here trade executed for 3 bid remaining 2 bid will be filled for next available maker
			// Now market price or latest trade price is 6
		} else {
			log.Fatal("unknown orderType and side combination")
		}
		//adjust size of maker order or delete maker order if size is zero
		// according to above example for this itteration fetched maker order has been settled and will be
		//deleted from order book
		err := makerDepth.decrSize(makerOrder.OrderId, size)
		if err != nil {
			log.Fatal(err)
		}

		// matched,write a log
		matchLog := newMatchLog(o.nextLogSeq(), o.product.Id, o.nextTradeSeq(), takerOrder, makerOrder, price, size)
		logs = append(logs, matchLog)
		fmt.Println("Last traded price ", price)
		// maker is filled
		if makerOrder.Size.IsZero() {

			doneLog := newDoneLog(o.nextLogSeq(), o.product.Id, makerOrder, makerOrder.Size, models.DoneReasonFilled)
			logs = append(logs, doneLog)
		}
	}

	//If pogram controller break out of loop
	//check if taker is of type limit and commodity to be trade is greater than 0
	if takerOrder.Type == models.OrderTypeLimit && takerOrder.Size.GreaterThan(decimal.Zero) {
		//It may be possible that there is no cross happened for entire order or
		//there was only partial order cross.
		//so it will be added to order book and set next log sequence in order to execute this order in future
		o.depths[takerOrder.Side].add(*takerOrder)
		openLog := newOpenLog(o.nextLogSeq(), o.product.Id, takerOrder)
		logs = append(logs, openLog)
	} else {
		//if marketorder and order dint execute cancel order
		//and if there is no more order to fullfill taker order and partial done

		//if takerorder is limit order and size<=0-->done
		//if takerorder is market order--->
		var remainingSize = takerOrder.Size
		var reason = models.DoneReasonFilled
		if takerOrder.Type == models.OrderTypeMarket {
			takerOrder.Price = decimal.Zero
			remainingSize = decimal.Zero
			if (takerOrder.Side == models.SideSell && takerOrder.Size.GreaterThan(decimal.Zero)) ||
				(takerOrder.Side == models.SideBuy && takerOrder.Funds.GreaterThan(decimal.Zero)) {
				reason = models.DoneReasonCancelled
			}
		}
		doneLog := newDoneLog(o.nextLogSeq(), o.product.Id, takerOrder, remainingSize, reason)
		logs = append(logs, doneLog)
	}
	return logs
}

func (o *orderBook) Snapshot() orderBookSnapshot {
	snapshot := orderBookSnapshot{
		Orders:        make([]BookOrder, len(o.depths[models.SideSell].orders)+len(o.depths[models.SideBuy].orders)),
		LogSeq:        o.logSeq,
		TradeSeq:      o.tradeSeq,
		OrderIdWindow: o.orderIdWindow,
	}

	i := 0
	for _, order := range o.depths[models.SideSell].orders {
		snapshot.Orders[i] = *order
		i++
	}

	for _, order := range o.depths[models.SideBuy].orders {
		snapshot.Orders[i] = *order
		i++
	}

	return snapshot
}

func (o *orderBook) Restore(snapshot *orderBookSnapshot) {
	o.logSeq = snapshot.LogSeq
	o.tradeSeq = snapshot.TradeSeq
	o.orderIdWindow = snapshot.OrderIdWindow
	if o.orderIdWindow.Cap == 0 {
		o.orderIdWindow = newWindow(0, orderIdWindowCap)
	}

	for _, order := range snapshot.Orders {
		o.depths[order.Side].add(order)
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

func NewOrderBook(product *models.Product) *orderBook {
	asks := &depth{
		queue:  treemap.NewWith(priceOrderIdKeyAscComparator),
		orders: map[int64]*BookOrder{},
	}

	bids := &depth{
		queue:  treemap.NewWith(priceOrderIdKeyDescComparator),
		orders: map[int64]*BookOrder{},
	}

	orderBook := &orderBook{
		product:       product,
		depths:        map[models.Side]*depth{models.SideBuy: bids, models.SideSell: asks},
		orderIdWindow: newWindow(0, orderIdWindowCap),
	}
	return orderBook
}
