/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/jhblack-olya/trade-engine/models"
)

type LogType string

const (
	LogTypeMatch   = LogType("match")
	LogTypeOpen    = LogType("open")
	LogTypeDone    = LogType("done")
	LogTypePending = LogType("pending")
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
	Art            int64
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
	Art            int64
	ExecutedValue  decimal.Decimal
	FilledSize     decimal.Decimal
	CancelledAt    string
	ExecutedAt     string
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
	TakerArt            int64
	MakerArt            int64
	TakerExecutedAt     string
	MakerExecutedAt     string
}

type PendingLog struct {
	Base
	OrderId       int64
	OrderType     int64
	Art           int64
	RemainingSize decimal.Decimal
	Side          models.Side
}

func newMatchLog(logSeq int64, productId string, tradeSeq int64, takerOrder, makerOrder *BookOrder, price, size decimal.Decimal, takerTimer, makerTimer int64, takerMatchedAt, makermatchedAt string) *MatchLog {
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
		TakerExecutedAt:     takerMatchedAt,
		MakerExecutedAt:     makermatchedAt,
	}
}
func newOpenLog(logSeq int64, productId string, takerOrder *BookOrder, timer int64) *OpenLog {
	fmt.Println("log seq achived by calling next log seq for open order ", logSeq)
	return &OpenLog{
		Base:           Base{LogTypeOpen, logSeq, productId, time.Now()},
		OrderId:        takerOrder.OrderId,
		RemainingSize:  takerOrder.Size,
		Price:          takerOrder.Price,
		Side:           takerOrder.Side,
		ExpiresIn:      timer,
		BackendOrderId: takerOrder.BackendOrderId,
	}
}

func newDoneLog(logSeq int64, productId string, order *BookOrder, remainingSize decimal.Decimal, reason models.DoneReason, timer int64, executedValue, filledSize decimal.Decimal, timeStamp string) *DoneLog {
	var (
		cancelledTime string
		matchedAt     string
	)

	if reason == models.DoneReasonCancelled {
		cancelledTime = timeStamp
	} else if reason == models.DoneReasonFilled || reason == models.DoneReasonPartial {
		matchedAt = timeStamp
	}
	return &DoneLog{
		Base:           Base{LogTypeDone, logSeq, productId, time.Now()},
		OrderId:        order.OrderId,
		Price:          order.Price,
		RemainingSize:  remainingSize,
		Reason:         reason,
		Side:           order.Side,
		ExpiresIn:      timer,
		BackendOrderId: order.BackendOrderId,
		ExecutedValue:  executedValue,
		FilledSize:     filledSize,
		CancelledAt:    cancelledTime,
		ExecutedAt:     matchedAt,
	}
}

func newPendingLog(logSeq int64, productId string, side models.Side, remainingSize decimal.Decimal, orderId, orderType int64) *PendingLog {
	return &PendingLog{
		Base:          Base{LogTypePending, logSeq, productId, time.Now()},
		OrderId:       orderId,
		RemainingSize: remainingSize,
		Side:          side,
		OrderType:     orderType,
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

func (l *PendingLog) GetSeq() int64 {
	return l.Sequence
}
