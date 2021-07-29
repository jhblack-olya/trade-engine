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
}

func newOpenLog(logSeq int64, productId string, takerOrder *BookOrder) *OpenLog {
	return &OpenLog{
		Base:          Base{LogTypeOpen, logSeq, productId, time.Now()},
		OrderId:       takerOrder.OrderId,
		RemainingSize: takerOrder.Size,
		Price:         takerOrder.Price,
		Side:          takerOrder.Side,
	}
}

func (l *OpenLog) GetSeq() int64 {
	return l.Sequence
}
