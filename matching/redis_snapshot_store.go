/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/jhblack-olya/trade-engine/conf"
	"github.com/jhblack-olya/trade-engine/models"
)

const (
	topicSnapshotPrefix = "matching_snapshot_"
)

type RedisSnapshotStore struct {
	productId   string
	redisClient *redis.Client
}

func NewRedisSnapshotStore(productId string) SnapshotStore {
	gbeConfig := conf.GetConfig()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     gbeConfig.Redis.Addr,
		Password: gbeConfig.Redis.Password,
		DB:       0,
	})

	return &RedisSnapshotStore{
		productId:   productId,
		redisClient: redisClient,
	}

}

func (s *RedisSnapshotStore) Store(snapshot *Snapshot) error {
	buf, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return s.redisClient.Set(topicSnapshotPrefix+s.productId, buf, 7*24*time.Hour).Err()
}

func (s *RedisSnapshotStore) GetLatest() (*Snapshot, error) {
	ret, err := s.redisClient.Get(topicSnapshotPrefix + s.productId).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			models.RedisErrCh <- err

		}
		return nil, err
	}
	var snapshot Snapshot
	err = json.Unmarshal(ret, &snapshot)
	return &snapshot, err
}
