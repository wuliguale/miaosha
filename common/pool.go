package common

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)


type Pool struct {
	mu sync.RWMutex
	channel chan *PoolConn
	config *PoolConfig
}


func NewPool(config *PoolConfig) (*Pool, error) {
	pool := &Pool{
		channel : make(chan *PoolConn, config.maxCap),
		config : config,
	}

	err := pool.InitPool()

	return pool, err
}


func (pool *Pool) InitPool() (err error) {
	pool.ClosePool()
	pool.channel = make(chan *PoolConn, pool.config.maxCap)

	for i := 0; i < pool.config.initCap; i++ {
		fmt.Println("newpool for ", i)

		poolConn, err := pool.makeConn()

		if err != nil {
			pool.ClosePool()
			return err
		}

		poolConn.updateIdleStartTime()
		pool.channel <- poolConn
	}

	return nil
}



func (pool *Pool) makeConn() (poolConn *PoolConn, err error) {
	serviceInfo, err := pool.config.GetServiceInfo()
	if err != nil {
		return nil, err
	}

	conn, err := pool.config.makeFunc(serviceInfo)

	if err != nil {
		return nil, err
	}

	poolConn = NewPoolConn(conn)
	return poolConn, nil
}


func (pool *Pool) Get() (closer io.Closer, err error) {
	pool.mu.RLock()
	channel := pool.channel
	config := pool.config
	pool.mu.RUnlock()

	if channel == nil {
		return nil, errors.New("pool channel empty")
	}

	//serviceInfoList is fast than pool.channel, update pool.channel
	if pool.config.serviceFast {
		err = pool.InitPool()
		if err != nil {
			return nil, err
		}

		pool.config.serviceFast = false
	}

	poolConn := &PoolConn{}

LOOP:
	select {
	case poolConn = <-channel:
		fmt.Println("pool get")
	default:
		poolConn, err = pool.makeConn()
		fmt.Println("pool get make")

		if err != nil {
			return nil, err
		}
	}

	if poolConn == nil {
		return nil, errors.New("poolConn is empty")
	}

	fmt.Println(poolConn.idleStartTime, time.Now().Unix(), config.maxIdleSeconds)

	//poolConn expired, get or make one
	if poolConn.idleStartTime > 0 && time.Now().Unix() - poolConn.idleStartTime >= config.maxIdleSeconds {
		pool.CloseConn(poolConn)

		goto LOOP
	}

	//validate
	if config.validateFunc != nil && !config.validateFunc(poolConn.conn) {
		pool.CloseConn(poolConn)

		goto LOOP
	}

	return poolConn.conn, nil
}


func (pool *Pool) Put(closer io.Closer) error {
	if closer == nil {
		return nil
	}

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	if pool.channel == nil {
		return pool.CloseConn(closer)
	}

	poolConn := NewPoolConn(closer)
	poolConn.updateIdleStartTime()

	select {
	case pool.channel <- poolConn:
		return nil
	default:
		// pool is full, close passed connection
		return pool.CloseConn(poolConn)
	}
}


func (pool *Pool) ClosePool() {
	pool.mu.Lock()

	channel := pool.channel

	pool.channel = nil
	pool.config = nil
	pool.mu.Unlock()

	if channel == nil {
		return
	}
	//close chan
	close(channel)

	for poolConn := range channel {
		pool.CloseConn(poolConn)
	}
}


func (pool *Pool) CloseConn(closer io.Closer) error {
	return closer.Close()
}

type PoolMakeFunc func(serviceInfo *ConsulServiceInfo) (io.Closer, error)
type PoolValidateFunc func(io.Closer) bool

type PoolConfig struct {
	initCap int
	maxCap int
	maxIdleSeconds int64
	serviceList *ConsulServiceInfoList
	serviceFast bool
	makeFunc PoolMakeFunc
	validateFunc PoolValidateFunc
	mu sync.RWMutex
}


func NewPoolConfig(initCap, maxCap int, maxIdleSeconds int64, serviceChan chan *ConsulServiceInfoList, makeFunc PoolMakeFunc, validateFunc PoolValidateFunc) (config *PoolConfig, err error) {
	if initCap < 0 || maxCap <= 0 || initCap > maxCap || maxIdleSeconds <= 0{
		return nil, errors.New("invalid pool config")
	}

	serviceInfoListInterface, err := SelectReceiveWithTimeout(serviceChan, 2)
	if err != nil {
		return nil, errors.New("pool get serviceInfoList from ch fail")
	}
	serviceInfoList, ok := serviceInfoListInterface.(*ConsulServiceInfoList)
	if !ok {
		return nil, errors.New("pool serviceInfoListInterface assert fail")
	}

	config = &PoolConfig{
		initCap:initCap,
		maxCap:maxCap,
		maxIdleSeconds:maxIdleSeconds,
		serviceList:serviceInfoList,
		serviceFast:false,
		makeFunc:makeFunc,
		validateFunc:validateFunc,
	}

	//update config service from serviceChan
	go func() {
		for {
			serviceInfoList2 , ok := <- serviceChan
			if !ok {
				//chan closed
				break
			}

			config.serviceList = serviceInfoList2
			config.serviceFast = true

			fmt.Println("serviceInfoList fast")
		}
	}()

	return config, nil
}


func (config *PoolConfig) GetServiceInfo() (serviceInfo *ConsulServiceInfo, err error) {
	return config.serviceList.GetNext()
}


type PoolConn struct {
	conn io.Closer
	mu sync.RWMutex
	idleStartTime int64
}


func NewPoolConn(closer io.Closer) *PoolConn{
	poolConn := &PoolConn{
		conn:closer,
		idleStartTime:0,
	}

	return poolConn
}


func (poolConn *PoolConn) Close() error {
	return poolConn.conn.Close()
}

func (poolConn *PoolConn) updateIdleStartTime() {
	if poolConn != nil {
		poolConn.idleStartTime = time.Now().Unix()
	}
}

