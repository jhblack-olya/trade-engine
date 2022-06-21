/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package models

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/go-redis/redis"
	"github.com/siddontang/go-log/log"

	"github.com/shopspring/decimal"

	"github.com/go-mysql-org/go-mysql/canal"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/utils"
)

type BinLogStream struct {
	canal.DummyEventHandler
	redisClient *redis.Client
}

func NewBinLogStream() *BinLogStream {
	gbeConfig := conf.GetConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     gbeConfig.Redis.Addr,
		Password: gbeConfig.Redis.Password,
		DB:       0,
	})

	return &BinLogStream{
		redisClient: redisClient,
	}
}

func (s *BinLogStream) OnRow(e *canal.RowsEvent) error {
	switch e.Table.Name {
	case "orderbooks": //"g_order":
		if e.Action == "delete" {
			return nil
		}

		var n = 0
		if e.Action == "update" {
			n = 1
		}

		var v Order
		s.parseRow(e, e.Rows[n], &v)

		buf, _ := json.Marshal(v)
		ret := s.redisClient.Publish(TopicOrder, buf)
		if ret.Err() != nil {
			log.Error(ret.Err())
		}

	case "g_account":
		var n = 0
		if e.Action == "update" {
			n = 1
		}

		var v Account
		s.parseRow(e, e.Rows[n], &v)

		buf, _ := json.Marshal(v)
		ret := s.redisClient.Publish(TopicAccount, buf)
		if ret.Err() != nil {
			log.Error(ret.Err())
		}

	case "g_fill":
		if e.Action == "delete" || e.Action == "update" {
			return nil
		}

		var v Fill
		s.parseRow(e, e.Rows[0], &v)

		buf, _ := json.Marshal(v)
		ret := s.redisClient.LPush(TopicFill, buf)
		if ret.Err() != nil {
			log.Error(ret.Err())
		}

	case "g_bill":
		if e.Action == "delete" || e.Action == "update" {
			return nil
		}

		var v Bill
		s.parseRow(e, e.Rows[0], &v)

		buf, _ := json.Marshal(v)
		ret := s.redisClient.LPush(TopicBill, buf)
		if ret.Err() != nil {
			log.Error(ret.Err())
		}
	}

	return nil
}

func (s *BinLogStream) parseRow(e *canal.RowsEvent, row []interface{}, dest interface{}) {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()
	num := v.NumField()
	for i := 0; i < num; i++ {
		f := v.Field(i)
		colName := t.Field(i).Name
		col := colName
		if e.Table.Name == "orderbooks" {
			switch colName {
			case "UserId":
				col = "user"
			case "ArtName":
				col = "art"
			case "BackendOrderId":
				col = "orderId"
			case "Size":
				col = "artBits"
			case "Funds":
				col = "totalAmount"
			case "FilledSize":
				col = "filledArtBits"
			case "ExecutedValue":
				col = "filledAmount"
			case "FillFees":
				col = "commission"
			case "Type":
				col = "orderType"
			case "CancelledAt":
				col = "cancelledAt"
			case "ExecutedAt":
				col = "executedAt"
			case "DeletedAt":
				col = "deletedAt"
			case "UserRole":
				col = "userRole"
			case "CommissionPercent":
				col = "commissionPercent"
			default:
				col = utils.SnakeCase(colName)
			}
		} else {
			col = utils.SnakeCase(colName)

		}
		colIdx := s.getColumnIndexByName(e, col)
		if e.Table.Name == "g_fill" && t.Field(i).Name == "ClientOid" {
			continue
		}
		if e.Table.Name == "orderbooks" && t.Field(i).Name == "CancelledAt" || t.Field(i).Name == "ExecutedAt" || t.Field(i).Name == "DeletedAt" {
			continue
		}
		rowVal := row[colIdx]

		switch f.Type().Name() {
		case "int64":
			f.SetInt(rowVal.(int64))
		case "string":
			f.SetString(rowVal.(string))
		case "bool":
			if rowVal.(int8) == 0 {
				f.SetBool(false)
			} else {
				f.SetBool(true)
			}
		case "Time":
			if rowVal != nil {
				f.Set(reflect.ValueOf(rowVal.(time.Time)))
			}
		case "Decimal":
			d := decimal.NewFromFloat(rowVal.(float64))
			f.Set(reflect.ValueOf(d))
		default:
			f.SetString(rowVal.(string))
		}
	}
}

func (s *BinLogStream) getColumnIndexByName(e *canal.RowsEvent, name string) int {
	for id, value := range e.Table.Columns {
		if value.Name == name {
			return id
		}
	}
	return -1
}

func (s *BinLogStream) Start() {
	gbeConfig := conf.GetConfig()

	cfg := canal.NewDefaultConfig()
	cfg.Addr = gbeConfig.DataSource.Addr
	cfg.User = gbeConfig.DataSource.User
	cfg.Password = gbeConfig.DataSource.Password
	cfg.Dump.ExecutionPath = ""
	cfg.Dump.TableDB = gbeConfig.DataSource.Database
	cfg.ParseTime = true
	cfg.IncludeTableRegex = []string{gbeConfig.DataSource.Database + "\\..*"}
	cfg.ExcludeTableRegex = []string{"mysql\\..*"}
	c, err := canal.NewCanal(cfg)
	if err != nil {
		MysqlErrCh <- err
		panic(err)
	}
	c.SetEventHandler(s)

	pos, err := c.GetMasterPos()
	if err != nil {
		MysqlErrCh <- err
		panic(err)
	}
	err = c.RunFrom(pos)
	if err != nil {
		MysqlErrCh <- err
		panic(err)
	}
}
