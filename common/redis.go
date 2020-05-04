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


func NewRedisPool() (redisPool *RedisPool, err error) {
	serviceInfo := &ConsulServiceInfo{
		Id:"1",
		Name:"redis-test",
		Host:"121.36.61.156",
		Port:6379,
	}

	serviceInfoList := &ConsulServiceInfoList{Name:"redis-test"}
	serviceInfoList.Add(serviceInfo)

	serviceChan := make(chan *ConsulServiceInfoList, 1)
	serviceChan <- serviceInfoList

	makeFunc := func(serviceInfo *ConsulServiceInfo) (io.Closer, error) {
		url := fmt.Sprintf("%s:%d", serviceInfo.Host, serviceInfo.Port)
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

	poolConfig, err := NewPoolConfig(3, 5, 60, serviceChan, makeFunc, validateFunc)
	if err != nil {
		return nil, err
	}
	pool, err :=  NewPool(poolConfig)
	if err != nil {
		return nil, err
	}

	redisPool = &RedisPool{pool}
	return redisPool, err
}


