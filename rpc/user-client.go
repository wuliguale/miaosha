package user

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"miaosha-demo/rpc/gen-go/user"
)

//rpc UserService的客户端
var defaultCtx = context.Background()


func RpcUserServiceReg(ip string, port int, userName, nickName, password string) (*user.UserStruct, error) {
	client, transport := GetRpcUserServiceClient(ip, port)
	defer transport.Close()

	return client.Reg(defaultCtx, userName, nickName, password)
}


func RpcUserServiceLogin(ip string, port int, userName, password string) (*user.UserStruct, error) {
	client, transport := GetRpcUserServiceClient(ip, port)
	defer transport.Close()

	return client.Login(defaultCtx, userName, password)
}


func GetRpcUserServiceClient(ip string, port int) (*user.UserServiceClient, thrift.TTransport){
	addr := fmt.Sprintf("%s:%d", ip, port)

	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transportFactory := thrift.NewTTransportFactory()

	//TSocket实现了TTransport接口
	socket, err := thrift.NewTSocket(addr)
	if err != nil {
		fmt.Println("Error opening socket:", err)
	}

	transport, err := transportFactory.GetTransport(socket)
	if err != nil {
		fmt.Println(err)
	}

	if err := transport.Open(); err != nil {
		fmt.Println(err)
	}

	//client拥有iprot+oprot，protocol拥有transport
	return user.NewUserServiceClientFactory(transport, protocolFactory), transport
}




