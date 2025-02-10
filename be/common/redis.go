package common

import "github.com/go-redis/redis/v8"

var RedisClient *redis.Client

func InitRedis(addr string, password string) *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	return RedisClient
}
