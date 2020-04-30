package main

import (
	"fmt"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"os/exec"
	"time"
)

func main() {

	out, err := exec.Command("/bin/bash", "-c", "ip addr show | grep 192.168.125.128").Output()

	fmt.Println(out, err)
	if err == nil && len(out) > 0 {
		fmt.Println("aaa")
	}

	fmt.Println("bbb")
	return


	config, err := common.NewConfigConsul()
	fmt.Println(err)

	freeCache := common.NewFreeCacheClient(50)

	consulClient, err := common.NewConsulClient(config, freeCache)

	mysqlPool, err := common.NewMysqlPool(consulClient)

	time.Sleep(time.Second * 60)

	conn, err := mysqlPool.Get()

	product := datamodels.Product{}
	res := conn.First(&product)
	fmt.Println(product, res.Error)

	time.Sleep(time.Second * 60)


	return
	/*
	redisPool, err := common.NewRedisPool()
	fmt.Println("start", err)

	conn, err := redisPool.Get()
	if err != nil {
		fmt.Println("pool get error,", err)
		return
	}

	username, err := redis.String(conn.Do("GET", "a"))
	fmt.Println(username, err)

	redisPool.Put(conn)


	time.Sleep(time.Second * 20)

	conn, err = redisPool.Get()
	username, err = redis.String(conn.Do("GET", "a"))
	fmt.Println(username, err)
	*/

}
