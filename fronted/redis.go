package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"miaosha-demo/common"
)

func main() {
	config, err := common.NewConfigConsul()
	if err != nil {
		fmt.Println(err)
		return
	}

	freeCache := common.NewFreeCacheClient(10)
	consul, err := common.NewConsulClient(config, freeCache)
	if err != nil {
		fmt.Println(err)
		return
	}

	redisClusterClient, err := common.NewRedisClusterClient(consul)

	pidOverKey := fmt.Sprintf("pid_over_%d", 6)
	isOver, err := redisClusterClient.Get(pidOverKey).Int()
	fmt.Println(isOver, err, redis.Nil == err)


}


