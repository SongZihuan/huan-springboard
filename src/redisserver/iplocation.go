package redisserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/api/apiip"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"time"
)

func QueryIpLocation(ip string) (*apiip.QueryIpLocationData, error) {
	key := fmt.Sprintf("ip:location:%s", ip)

	cacheRes := func() *apiip.QueryIpLocationData {
		res, err := rdb.Get(context.Background(), key).Result()
		if err != nil {
			return nil
		}

		var loc apiip.QueryIpLocationData
		err = json.Unmarshal([]byte(res), &loc)
		if err != nil {
			return nil
		}

		return &loc
	}()
	if cacheRes != nil {
		return cacheRes, nil
	}

	res, err := apiip.QueryIpLocation(ip)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	err = rdb.Set(context.Background(), key, string(data), 24*time.Hour).Err()
	if err != nil {
		logger.Errorf("redis set error: %s", err.Error())
	}

	return res, nil
}
