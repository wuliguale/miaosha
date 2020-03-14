package user

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"miaosha-demo/rpc/gen-go/user"
)

//rpc UserService的客户端
var defaultCtx = context.Background()

func RpcUserServiceReg(userName, nickName, password string) (*user.UserStruct, error) {
	client, transport := GetRpcUserServiceClient()
	defer transport.Close()

	return client.Reg(defaultCtx, userName, nickName, password)
}


func RpcUserServiceLogin(userName, password string) (*user.UserStruct, error) {
	client, transport := GetRpcUserServiceClient()
	defer transport.Close()

	return client.Login(defaultCtx, userName, password)
}


func GetRpcUserServiceClient() (*user.UserServiceClient, thrift.TTransport){
	addr := "127.0.0.1:9090"

	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transportFactory := thrift.NewTTransportFactory()

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

	return user.NewUserServiceClientFactory(transport, protocolFactory), transport
}




