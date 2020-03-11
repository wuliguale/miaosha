package common

import (
	"github.com/garyburd/redigo/redis"
)

func NewRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:10,
		MaxActive:10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "192.168.125.128:6379")
			return conn, err
		},
	}
}
