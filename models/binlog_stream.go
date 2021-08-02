package models

import (
	"github.com/go-redis/redis"
	"github.com/siddontang/go-mysql/canal"
	"gitlab.com/gae4/trade-engine/conf"
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
		panic(err)
	}
	c.SetEventHandler(s)

	pos, err := c.GetMasterPos()
	if err != nil {
		panic(err)
	}
	err = c.RunFrom(pos)
	if err != nil {
		panic(err)
	}
}
