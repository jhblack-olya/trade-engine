/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package worker

import (
	"time"

	"github.com/prometheus/common/log"
	"github.com/jhblack-olya/trade-engine/matching"
	"github.com/jhblack-olya/trade-engine/models"
	"github.com/jhblack-olya/trade-engine/models/mysql"
	"github.com/jhblack-olya/trade-engine/service"
)

type TradeMaker struct {
	tradeCh   chan *models.Trade
	logReader matching.LogReader
	logOffset int64
	logSeq    int64
}

//NewTradeMaker: pushes fullfilled trade into database log stream
func NewTradeMaker(logReader matching.LogReader) *TradeMaker {
	t := &TradeMaker{
		tradeCh:   make(chan *models.Trade, 1000),
		logReader: logReader,
	}

	lastTrade, err := mysql.SharedStore().GetLastTradeByProductId(logReader.GetProductId())
	if err != nil {
		panic(err)
	}
	if lastTrade != nil {
		t.logOffset = lastTrade.LogOffset
		t.logSeq = lastTrade.LogSeq
	}

	t.logReader.RegisterObserver(t)
	return t
}

func (t *TradeMaker) Start() {
	if t.logOffset > 0 {
		t.logOffset++
	}
	go t.logReader.Run(t.logSeq, t.logOffset)
	go t.runFlusher()
}

func (t *TradeMaker) OnOpenLog(log *matching.OpenLog, offset int64) {
	// do nothing
}

func (t *TradeMaker) OnDoneLog(log *matching.DoneLog, offset int64) {
	// do nothing
}

func (t *TradeMaker) OnMatchLog(log *matching.MatchLog, offset int64) {
	t.tradeCh <- &models.Trade{
		Id:              log.TradeId,
		ProductId:       log.ProductId,
		TakerOrderId:    log.TakerOrderId,
		MakerOrderId:    log.MakerOrderId,
		Price:           log.Price,
		Size:            log.Size,
		Side:            log.Side,
		Time:            log.Time,
		LogOffset:       offset,
		LogSeq:          log.Sequence,
		TakerArt:        log.TakerArt,
		MakerArt:        log.MakerArt,
		TakerExecutedAt: log.TakerExecutedAt,
		MakerExecutedAt: log.MakerExecutedAt,
	}
}

func (t *TradeMaker) runFlusher() {
	var trades []*models.Trade

	for {
		select {
		case trade := <-t.tradeCh:
			trades = append(trades, trade)

			if len(t.tradeCh) > 0 && len(trades) < 1000 {
				continue
			}

			// Ensure successful storage
			for {
				err := service.AddTrades(trades)
				if err != nil {
					log.Error(err)
					time.Sleep(time.Second)
					continue
				}
				trades = nil
				break
			}
		}
	}
}
