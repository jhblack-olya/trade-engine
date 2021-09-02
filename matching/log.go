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
	OrderId        int64
	RemainingSize  decimal.Decimal
	Price          decimal.Decimal
	Side           models.Side
	ExpiresIn      int64
	BackendOrderId string
	Art            string
}

type DoneLog struct {
	Base
	OrderId        int64
	Price          decimal.Decimal
	RemainingSize  decimal.Decimal
	Reason         models.DoneReason
	Side           models.Side
	ExpiresIn      int64
	BackendOrderId string
	Art            string
}

type MatchLog struct {
	Base
	TradeId             int64
	TakerOrderId        int64
	MakerOrderId        int64
	Side                models.Side
	Price               decimal.Decimal
	Size                decimal.Decimal
	TakerClientOid      string
	MakerClientOid      string
	TakerExpiresIn      int64
	MakerExpiresIn      int64
	TakerBackendOrderId string
	MakerBackendOrderId string
	TakerArt            string
	MakerArt            string
}

func newMatchLog(logSeq int64, productId string, tradeSeq int64, takerOrder, makerOrder *BookOrder, price, size decimal.Decimal, takerTimer, makerTimer int64, takerArt, makerArt string) *MatchLog {
	return &MatchLog{
		Base:                Base{LogTypeMatch, logSeq, productId, time.Now()},
		TradeId:             tradeSeq,
		TakerOrderId:        takerOrder.OrderId,
		MakerOrderId:        makerOrder.OrderId,
		Side:                makerOrder.Side,
		Price:               price,
		Size:                size,
		TakerClientOid:      takerOrder.ClientOid,
		MakerClientOid:      makerOrder.ClientOid,
		TakerExpiresIn:      takerTimer,
		MakerExpiresIn:      makerTimer,
		TakerBackendOrderId: takerOrder.BackendOrderId,
		MakerBackendOrderId: makerOrder.BackendOrderId,
		TakerArt:            takerArt,
		MakerArt:            makerArt,
	}
}
func newOpenLog(logSeq int64, productId string, takerOrder *BookOrder, timer int64, art string) *OpenLog {
	return &OpenLog{
		Base:           Base{LogTypeOpen, logSeq, productId, time.Now()},
		OrderId:        takerOrder.OrderId,
		RemainingSize:  takerOrder.Size,
		Price:          takerOrder.Price,
		Side:           takerOrder.Side,
		ExpiresIn:      timer,
		BackendOrderId: takerOrder.BackendOrderId,
		Art:            art,
	}
}

func newDoneLog(logSeq int64, productId string, order *BookOrder, remainingSize decimal.Decimal, reason models.DoneReason, timer int64, art string) *DoneLog {
	return &DoneLog{
		Base:           Base{LogTypeDone, logSeq, productId, time.Now()},
		OrderId:        order.OrderId,
		Price:          order.Price,
		RemainingSize:  remainingSize,
		Reason:         reason,
		Side:           order.Side,
		ExpiresIn:      timer,
		BackendOrderId: order.BackendOrderId,
		Art:            art,
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
