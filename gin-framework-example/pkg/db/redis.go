package db

import (
	"context"
	"fmt"

	"gin-framework-example/internal/conf"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

func InitRedis() {
	redisCfg := conf.Conf.Redis
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	_, err := Rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}
