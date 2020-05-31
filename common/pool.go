package common

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
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
	//clear chan and close conn
	for {
		isBreak := false

		select {
		case poolConn := <- pool.channel:
			pool.CloseConn(poolConn)
		default:
			isBreak = true
		}

		if isBreak {
			break
		}
	}

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

	//get conn from chan
	breakNow := false
	for {
		select {
		case poolConn = <-channel:
			if pool.Validate(poolConn) {
				zap.L().Info("get conn from  chan succ")
				breakNow = true
			} else {
				pool.CloseConn(poolConn)
				poolConn.conn = nil

				zap.L().Info("get conn from chan validate fail")
			}
		//timeout
		case <-time.After(time.Millisecond * 500):
			zap.L().Info("get conn from chan timeout")
			breakNow = true
		}

		if breakNow {
			break
		}
	}
	
	//new conn
	if poolConn.conn == nil {
		poolConn, err = pool.makeConn()
		if err != nil {
			zap.L().Info("pool new conn fail")
			return nil, err
		}

		if !pool.Validate(poolConn) {
			pool.CloseConn(poolConn)

			zap.L().Info("pool new conn validate fail")
			return nil, errors.New("pool conn validate fail")
		}
	}

	return poolConn.conn, nil
}


func (pool *Pool) Validate(poolConn *PoolConn) bool {
	pool.mu.RLock()
	config := pool.config
	pool.mu.RUnlock()

	//poolConn expired
	if poolConn.idleStartTime > 0 && time.Now().Unix() - poolConn.idleStartTime >= config.maxIdleSeconds {
		return false
	}

	//validate
	if config.validateFunc != nil && !config.validateFunc(poolConn.conn) {
		return false
	}

	return true
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

	var serviceInfoList *ConsulServiceInfoList

	select {
	case serviceInfoList2, ok := <- serviceChan:
		if !ok {
			//chan closed
			return nil, errors.New("pool get serviceInfoList from ch close")
		}
		serviceInfoList = serviceInfoList2
	case <-time.After(time.Second * 2):
		return nil, errors.New("pool get serviceInfoList from ch timeout")
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
			time.Sleep(time.Second * 1)
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

