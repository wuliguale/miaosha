package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"miaosha-demo/fronted/web/controllers"
	"strconv"
)

func main() {
	redisConn := controllers.RedisPool.Get()
	defer redisConn.Close()

	pid := 1
	key := "pid_over_" + strconv.Itoa(int(pid))
	res, err := redis.Int(redisConn.Do("get", key))

	fmt.Println(key)
	fmt.Println(res)
	fmt.Print(err, err == redis.ErrNil)

}