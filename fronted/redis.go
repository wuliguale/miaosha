package main

import (
	"fmt"
	"os"

	//"github.com/garyburd/redigo/redis"
	//"miaosha-demo/fronted/web/controllers"
	//"strconv"
	"time"
)

import redis2 "github.com/go-redis/redis/v7"

func main() {
	path, err := os.Getwd()
	fmt.Println(path, err)
	return
	
	ExampleClient()

	return


	/*
	redisConn := controllers.RedisPool.Get()
	defer redisConn.Close()

	pid := 1
	key := "pid_over_" + strconv.Itoa(int(pid))
	res, err := redis.Int(redisConn.Do("get", key))

	fmt.Println(key)
	fmt.Println(res)
	fmt.Print(err, err == redis.ErrNil)
*/




}


func ExampleClient() {
	client := redis2.NewClusterClient(&redis2.ClusterOptions{
		Addrs:    []string{"121.36.61.156:6384"},
		Password: "310900",
	})

	a, err := client.Get("a").Result()
	fmt.Println(a, err)
	return


	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	fmt.Println("pool state init state:", client.PoolStats())

	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("key:%d", i)
		v := k
		val, err := client.Set(k, v, 60*time.Second).Result()
		if err != nil {
			panic(err)
		}

		val, err = client.Get(k).Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("key:", val)
	}
	fmt.Println("pool state final state:", client.PoolStats()) //获取客户端连接池相关信息
}