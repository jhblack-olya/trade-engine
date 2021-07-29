package matching

import (
	"math"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/models"
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
		// if both false break
		if (takerOrder.Side == models.SideBuy && takerOrder.Price.LessThan(makerOrder.Price)) ||
			(takerOrder.Side == models.SideSell && takerOrder.Price.GreaterThan(makerOrder.Price)) {
			break
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
	}
	return logs
}
