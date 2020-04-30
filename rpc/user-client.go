package user

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"io"
	"miaosha-demo/common"
	"miaosha-demo/rpc/gen-go/user"
	"reflect"
	"strconv"
)

//rpc UserService的客户端
var defaultCtx = context.Background()


func (rpcUser *RpcUser) CallReg(userName, nickName, password string) (*user.UserStruct, error) {
	return rpcUser.Call("Reg",userName, nickName, password)
}

func (rpcUser *RpcUser) RpcUserServiceLogin(userName, password string) (*user.UserStruct, error) {
	return rpcUser.Call("Login", userName, password)
}


type RpcUser struct {
	transportPool *common.Pool
	protocolFactory thrift.TProtocolFactory
}


func NewRpcUser (consul *common.ConsulClient) (rpcUser *RpcUser, err error) {
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	pool, err := NewTransportPool(consul)
	if err != nil {
		return nil, err
	}

	rpcUser = &RpcUser{
		transportPool:pool,
		protocolFactory:protocolFactory,
	}

	return rpcUser, nil
}


func NewTransportPool(consul *common.ConsulClient) (pool *common.Pool, err error) {
	serviceName := "miaosha-demo-rpc-user"
	serviceInfoList, err := consul.GetServiceListByName(serviceName)
	if err != nil {
		return nil, err
	}

	var addressList []map[string]string
	for _,serviceInfo := range serviceInfoList.List {
		address := map[string]string{
			"host" : serviceInfo.Host,
			"port" : strconv.Itoa(serviceInfo.Port),
		}

		addressList = append(addressList, address)
	}

	fmt.Println(addressList)

	makeFunc := func(address map[string]string) (io.Closer, error) {
		addr := fmt.Sprintf("%s:%d", address["host"], address["port"])
		transportFactory := thrift.NewTTransportFactory()

		//TSocket实现了TTransport接口
		socket, err := thrift.NewTSocket(addr)
		if err != nil {
			return nil, err
		}

		transport, err := transportFactory.GetTransport(socket)
		if err := transport.Open(); err != nil {
			fmt.Println(err)
		}

		return transport, nil
	}

	poolConfig, err := common.NewPoolConfig(1, 2, 3, addressList, 0, makeFunc, nil)
	return  common.NewPool(poolConfig)
}


func (rpcUser *RpcUser) Call(method string, vals ...string) (userStruct *user.UserStruct, err error) {
	closer, err := rpcUser.transportPool.Get()
	fmt.Println(err)

	transport, ok := closer.(thrift.TTransport)
	if !ok {

	}

	//client拥有iprot+oprot，protocol拥有transport
	client := user.NewUserServiceClientFactory(transport, rpcUser.protocolFactory)

	//call
	in := []reflect.Value{}
	in = append(in, reflect.ValueOf(defaultCtx))

	for _, v := range vals {
		in = append(in, reflect.ValueOf(v))
	}
	abc := reflect.ValueOf(client).MethodByName(method).Call(in)
	userStruct = abc[0].Interface().(*user.UserStruct)
	//return client.Reg(defaultCtx, userName, nickName, password)

	rpcUser.transportPool.Put(transport)

	return userStruct, nil
}






