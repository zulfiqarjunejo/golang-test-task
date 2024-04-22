package redis

import (
	"github.com/go-redis/redis"
)

func NewRedis() (*redis.Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := r.Ping().Result()
	if err != nil {
		return nil, err
	}

	return r, nil
}
