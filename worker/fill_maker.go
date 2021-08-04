package worker

import (
	"fmt"

	"github.com/pingcap/log"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
	"gitlab.com/gae4/trade-engine/service"
)

type FillMaker struct {
	fillCh    chan *models.Fill
	logReader matching.LogReader
	logOffset int64
	logSeq    int64
}

func NewFillMaker(logReader matching.LogReader) *FillMaker {
	t := &FillMaker{
		fillCh:    make(chan *models.Fill, 1000),
		logReader: logReader,
	}

	lastFill, err := mysql.SharedStore().GetLastFillByProductId(logReader.GetProductId())
	if err != nil {
		panic(err)
	}
	if lastFill != nil {
		t.logOffset = lastFill.LogOffset
		t.logSeq = lastFill.LogSeq
	}
	fmt.Printf("initial logSeq %v at logOffset %v", t.logSeq, t.logOffset)
	t.logReader.RegisterObserver(t)
	return t
}

func (t *FillMaker) Start() {
	if t.logOffset > 0 {
		t.logOffset++
	}
	fmt.Printf("logSeq on Start() %v at offset %v", t.logSeq, t.logOffset)
	go t.logReader.Run(t.logSeq, t.logOffset)
	go t.flusher()
}

func (t *FillMaker) OnOpenLog(log *matching.OpenLog, offset int64) {
	_, _ = service.UpdateOrderStatus(log.OrderId, models.OrderStatusNew, models.OrderStatusOpen)
}

func (t *FillMaker) OnMatchLog(log *matching.MatchLog, offset int64) {
	t.fillCh <- &models.Fill{
		TradeId:    log.TradeId,
		MessageSeq: log.Sequence,
		OrderId:    log.TakerOrderId,
		ProductId:  log.ProductId,
		Size:       log.Size,
		Price:      log.Price,
		Liquidity:  "T",
		Side:       log.Side,
		LogOffset:  offset,
		LogSeq:     log.Sequence,
	}
	t.fillCh <- &models.Fill{
		TradeId:    log.TradeId,
		MessageSeq: log.Sequence,
		OrderId:    log.MakerOrderId,
		ProductId:  log.ProductId,
		Size:       log.Size,
		Price:      log.Price,
		Liquidity:  "M",
		Side:       log.Side.Opposite(),
		LogOffset:  offset,
		LogSeq:     log.Sequence,
	}
}

func (t *FillMaker) OnDoneLog(log *matching.DoneLog, offset int64) {
	t.fillCh <- &models.Fill{
		MessageSeq: log.Sequence,
		OrderId:    log.OrderId,
		ProductId:  log.ProductId,
		Size:       log.RemainingSize,
		Done:       true,
		DoneReason: log.Reason,
		LogOffset:  offset,
		LogSeq:     log.Sequence,
	}
}

func (t *FillMaker) flusher() {
	var fills []*models.Fill
	for {
		select {
		case fill := <-t.fillCh:
			fills = append(fills, fill)
			fills = append(fills, fill)
			if len(t.fillCh) > 0 && len(fills) < 1000 {
				continue
			}

			for {
				err := service.AddFills(fills)
				if err != nil {
					log.Error(err.Error())
					continue
				}
				fills = nil
				break
			}
		}
	}
}
