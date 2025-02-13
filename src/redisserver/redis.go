package redisserver

import (
	"context"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis() error {
	if rdb != nil {
		return nil
	}

	if !config.IsReady() {
		panic("config not ready")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     config.GetConfig().Redis.Address,
		Password: config.GetConfig().Redis.Password, // no password set
		DB:       config.GetConfig().Redis.DB,       // use default DB
	})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return err
	}

	return nil
}

func CloseRedis() {
	if rdb == nil {
		return
	}

	_ = rdb.Close()
	rdb = nil
}
