package redisserver

import (
	"context"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"time"
)

const BannedData = "banned"

func SetSSHIpBanned(ip string, ttl time.Duration) error {
	key := fmt.Sprintf("ssh:ip:banned:%s", ip)

	res1, err := rdb.TTL(context.Background(), key).Result()
	if err != nil {
		return err
	} else if res1 == -1 { // ip被设置封禁且没有TTL
		logger.Warnf("ip: %s is banned by redis forver", ip)
		return nil
	} else if res1 > ttl { // 原封禁时长更长，则不做变化
		return nil
	}

	_, err = rdb.Set(context.Background(), key, BannedData, ttl).Result()
	if err != nil {
		return err
	}

	return nil
}

func QuerySSHIpBanned(ip string) bool { // 返回 true 表示放行
	key := fmt.Sprintf("ssh:ip:banned:%s", ip)

	res1, err := rdb.TTL(context.Background(), key).Result()
	if err != nil {
		logger.Warnf("query ssh ip (%s) banned from redis error: %s", ip, err.Error())
		return false
	} else if res1 == -1 { // ip被设置封禁且没有TTL
		logger.Warnf("ip: %s is banned by redis forver", ip)
		return false
	} else if res1 == -2 { // 键不存在
		return true
	} else { // 键存在且有设置TTL
		return false
	}
}
