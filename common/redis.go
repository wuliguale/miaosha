package common

import (
	"fmt"
	redis2 "github.com/garyburd/redigo/redis"
	"github.com/go-redis/redis/v7"
	"io"
	"strconv"
)

func NewRedisClusterClient(consul *ConsulClient) (*redis.ClusterClient, error) {
	serviceName := consul.Config.GetRedisServiceName()
	serviceInfoList, err := consul.GetServiceListByName(serviceName)
	if err != nil {
		return nil, err
	}

	var addrList []string
	for _,serviceInfo := range serviceInfoList.List {
		addrList = append(addrList, serviceInfo.Host + ":" + strconv.Itoa(serviceInfo.Port))
	}

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs : addrList,
		Password: "310900",
	}), nil
}


type RedisPool struct {
	pool *Pool
}

func (redisPool RedisPool) Get() (conn redis2.Conn, err error) {
	closer, err := redisPool.pool.Get()
	if err != nil {
		fmt.Println("get error", err)
		return
	}

	conn, ok := closer.(redis2.Conn)
	if !ok {
		fmt.Println("assert", ok)
	}

	return conn, nil
}


func (redisPool RedisPool) Put(conn redis2.Conn) error {
	return redisPool.pool.Put(conn)
}


func (redisPool RedisPool) Close(conn redis2.Conn) error {
	return redisPool.pool.CloseConn(conn)
}


func NewRedisPool() (redisPool RedisPool, err error) {
	addressList := []map[string]string{
		{
			"host" : "121.36.61.156",
			"port" : "6379",
		},
	}

	makeFunc := func(address map[string]string) (io.Closer, error) {
		url := fmt.Sprintf("%s:%s", address["host"], address["port"])
		return redis2.Dial("tcp", url)
	}

	validateFunc := func(closer io.Closer) bool {
		conn, ok := closer.(redis2.Conn)
		fmt.Println(ok)

		username, err := redis2.String(conn.Do("ping"))
		fmt.Println(username, err)

		if username == "PONG" {
			return true
		} else {
			return false
		}
	}

	poolConfig, err := NewPoolConfig(3, 5, 60, addressList, 0, makeFunc, validateFunc)
	pool, err :=  NewPool(poolConfig)

	redisPool = RedisPool{pool}
	return redisPool, err
}


