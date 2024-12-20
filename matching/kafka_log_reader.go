/*
	Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.

You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package matching

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	logger "github.com/siddontang/go-log/log"
	"github.com/jhblack-olya/trade-engine/models"
)

type KafkaLogReader struct {
	readerId  string
	productId string
	reader    *kafka.Reader
	observer  LogObserver
}

func NewKafkaLogReader(readerId, productId string, brokers []string) LogReader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topicBookMessagePrefix + productId,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	return &KafkaLogReader{readerId: readerId, productId: productId, reader: reader}
}

func (r *KafkaLogReader) GetProductId() string {
	return r.productId
}
func (r *KafkaLogReader) RegisterObserver(observer LogObserver) {
	r.observer = observer
}

func (r *KafkaLogReader) Run(seq, offset int64) {
	logger.Infof("%v:%v read from %v", r.productId, r.readerId, offset)

	var lastSeq = seq

	err := r.reader.SetOffset(offset)
	if err != nil {
		panic(err)
	}

	for {
		kMessage, err := r.reader.FetchMessage(context.Background())
		if err != nil {
			logger.Error(err)
			models.KafkaErrCh <- err

			continue
		}

		var base Base
		err = json.Unmarshal(kMessage.Value, &base)
		if err != nil {
			panic(err)
		}
		fmt.Println("base.Sequnce = ", base.Sequence, " from kafka log")
		fmt.Println("last Sequnce = ", lastSeq, " from last fill")

		if base.Sequence <= lastSeq {
			logger.Info("%v:%v discard log :%+v", r.productId, r.readerId, base)
			continue
		} else if lastSeq > 0 && base.Sequence != lastSeq+1 {
			logger.Fatalf("non-sequence detected, lastSeq=%v seq=%v", lastSeq, base.Sequence)
		}

		lastSeq = base.Sequence

		switch base.Type {
		case LogTypeOpen:
			var log OpenLog
			err := json.Unmarshal(kMessage.Value, &log)
			if err != nil {
				panic(err)
			}
			r.observer.OnOpenLog(&log, kMessage.Offset)
		case LogTypeMatch:
			var log MatchLog
			err := json.Unmarshal(kMessage.Value, &log)
			if err != nil {
				panic(err)
			}
			r.observer.OnMatchLog(&log, kMessage.Offset)
		case LogTypeDone:
			var log DoneLog
			err := json.Unmarshal(kMessage.Value, &log)
			if err != nil {
				panic(err)
			}
			r.observer.OnDoneLog(&log, kMessage.Offset)

		}
	}
}
