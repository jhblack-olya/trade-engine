/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package mysql

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/mysql"
	"github.com/prometheus/common/log"
	"github.com/jhblack-olya/trade-engine/conf"
	"github.com/jhblack-olya/trade-engine/models"
)

var gdb *gorm.DB
var store models.Store
var storeOnce sync.Once

type Store struct {
	db *gorm.DB
}

func SharedStore() models.Store {
	storeOnce.Do(func() {
		err := initDb()
		if err != nil {
			panic(err)
		}
		store = NewStore(gdb)
	})
	return store
}

func NewStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

func initDb() error {
	cfg := conf.GetConfig()

	url := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
		cfg.DataSource.User, cfg.DataSource.Password, cfg.DataSource.Addr, cfg.DataSource.Database)
	var err error
	gdb, err = gorm.Open(cfg.DataSource.DriverName, url)
	if err != nil {
		return err
	}

	gdb.SingularTable(true)
	gdb.DB().SetMaxIdleConns(10)
	gdb.DB().SetMaxOpenConns(50)

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "g_" + defaultTableName
	}

	if cfg.DataSource.EnableAutoMigrate {
		var tables = []interface{}{
			&models.Account{},
			&models.Order{},
			&models.Product{},
			&models.Trade{},
			&models.Fill{},
			&models.User{},
			&models.Bill{},
			&models.Config{},
		}
		for _, table := range tables {
			log.Infof("migrating database, table: %v", reflect.TypeOf(table))
			if err = gdb.AutoMigrate(table).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) BeginTx() (models.Store, error) {
	db := s.db.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	return NewStore(db), nil
}

func (s *Store) Rollback() error {
	return s.db.Rollback().Error
}

func (s *Store) CommitTx() error {
	return s.db.Commit().Error
}
