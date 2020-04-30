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

	for i := 0; i < config.initCap; i++ {
		fmt.Println("newpool for ", i)

		poolConn, err := pool.makeConn()

		if err != nil {
			pool.ClosePool()
			return nil, err
		}

		poolConn.updateIdleStartTime()
		pool.channel <- poolConn
	}

	return pool, nil
}

func (pool *Pool) makeConn() (poolConn *PoolConn, err error) {
	address := pool.config.GetAddress()
	conn, err := pool.config.makeFunc(address)

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

type PoolMakeFunc func(map[string]string) (io.Closer, error)
type PoolValidateFunc func(io.Closer) bool

type PoolConfig struct {
	initCap int
	maxCap int
	maxIdleSeconds int64
	addressList []map[string]string
	addressIndex int
	makeFunc PoolMakeFunc
	validateFunc PoolValidateFunc
	mu sync.RWMutex
}


func NewPoolConfig(initCap, maxCap int, maxIdleSeconds int64, addressList []map[string]string, addressIndex int, makeFunc PoolMakeFunc, validateFunc PoolValidateFunc) (config *PoolConfig, err error) {
	if initCap < 0 || maxCap <= 0 || initCap > maxCap || maxIdleSeconds <= 0{
		return nil, errors.New("invalid pool config")
	}

	config = &PoolConfig{
		initCap:initCap,
		maxCap:maxCap,
		maxIdleSeconds:maxIdleSeconds,
		addressList:addressList,
		addressIndex:addressIndex,
		makeFunc:makeFunc,
		validateFunc:validateFunc,
	}

	return config, nil
}


func (config *PoolConfig) GetAddress() (address map[string]string) {
	address = map[string]string{}

	len := len(config.addressList)
	if len > 0 {
		next := (config.addressIndex + 1) % len
		address = config.addressList[next]
		config.addressIndex = next
	}

	return address
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

