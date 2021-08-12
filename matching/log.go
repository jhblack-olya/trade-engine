package matching

import (
	"time"

	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/models"
)

type LogType string

const (
	LogTypeMatch = LogType("match")
	LogTypeOpen  = LogType("open")
	LogTypeDone  = LogType("done")
)

type Log interface {
	GetSeq() int64
}

type Base struct {
	Type      LogType
	Sequence  int64
	ProductId string
	Time      time.Time
}

type OpenLog struct {
	Base
	OrderId       int64
	RemainingSize decimal.Decimal
	Price         decimal.Decimal
	Side          models.Side
	ExpiresIn     int64
}

type DoneLog struct {
	Base
	OrderId       int64
	Price         decimal.Decimal
	RemainingSize decimal.Decimal
	Reason        models.DoneReason
	Side          models.Side
	ExpiresIn     int64
}

type MatchLog struct {
	Base
	TradeId        int64
	TakerOrderId   int64
	MakerOrderId   int64
	Side           models.Side
	Price          decimal.Decimal
	Size           decimal.Decimal
	TakerClientOid string
	MakerClientOid string
	TakerExpiresIn int64
	MakerExpiresIn int64
}

func newMatchLog(logSeq int64, productId string, tradeSeq int64, takerOrder, makerOrder *BookOrder, price, size decimal.Decimal, takerTimer, makerTimer int64) *MatchLog {
	return &MatchLog{
		Base:           Base{LogTypeMatch, logSeq, productId, time.Now()},
		TradeId:        tradeSeq,
		TakerOrderId:   takerOrder.OrderId,
		MakerOrderId:   makerOrder.OrderId,
		Side:           makerOrder.Side,
		Price:          price,
		Size:           size,
		TakerClientOid: takerOrder.ClientOid,
		MakerClientOid: makerOrder.ClientOid,
		TakerExpiresIn: takerTimer,
		MakerExpiresIn: makerTimer,
	}
}
func newOpenLog(logSeq int64, productId string, takerOrder *BookOrder, timer int64) *OpenLog {
	return &OpenLog{
		Base:          Base{LogTypeOpen, logSeq, productId, time.Now()},
		OrderId:       takerOrder.OrderId,
		RemainingSize: takerOrder.Size,
		Price:         takerOrder.Price,
		Side:          takerOrder.Side,
		ExpiresIn:     timer,
	}
}

func newDoneLog(logSeq int64, productId string, order *BookOrder, remainingSize decimal.Decimal, reason models.DoneReason, timer int64) *DoneLog {
	return &DoneLog{
		Base:          Base{LogTypeDone, logSeq, productId, time.Now()},
		OrderId:       order.OrderId,
		Price:         order.Price,
		RemainingSize: remainingSize,
		Reason:        reason,
		Side:          order.Side,
		ExpiresIn:     timer,
	}
}

func (l *OpenLog) GetSeq() int64 {
	return l.Sequence
}

func (l *DoneLog) GetSeq() int64 {
	return l.Sequence
}

func (l *MatchLog) GetSeq() int64 {
	return l.Sequence
}
