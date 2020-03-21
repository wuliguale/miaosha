package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"log"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
	"miaosha-demo/rpc/gen-go/user"
	"miaosha-demo/services"
)

//rpc UserService的服务端
type UserServiceHandler struct {
	UserService services.IUserService
	ConsulClient *common.ConsulClient
}

func NewUserServiceHandler() (userServiceHandler *UserServiceHandler, err error) {
	db, err := common.NewMysqlConn()
	if err != nil {
		log.Fatalln("prc user new  handler fail", err)
	}

	userService := services.NewUserService(repositories.NewUserRepository(db))

	consulConfig, err := common.NewConfigConsul()
	if err != nil {
		log.Fatalln("new config consul fail: ", err)
	}

	consulClient, err := common.NewConsulClient(consulConfig, nil)
	if err != nil {
		log.Fatalln("new consul client fail: ", err)
	}

	return &UserServiceHandler{userService, consulClient}, nil
}


func (h *UserServiceHandler) Reg(ctx context.Context, userName, nickName, password string) (*user.UserStruct, error) {
	modelUser := &datamodels.User{
		UserName: userName,
		NickName: nickName,
		Password: password,
	}

	err := h.UserService.InsertUser(modelUser)
	if err != nil {
		return &user.UserStruct{}, err
	}

	modelUser, err  = h.UserService.GetUserByName(userName)
	if err != nil {
		return &user.UserStruct{}, err
	}

	userSt := h.ModelUser2StructUset(modelUser)
	return userSt, nil
}


func (h *UserServiceHandler) Login(ctx context.Context, userName , password string) (*user.UserStruct, error) {
	modelUser, isOk := h.UserService.IsPwdSuccess(userName, password)
	if !isOk {
		return &user.UserStruct{}, errors.New("login fail")
	}

	userSt := h.ModelUser2StructUset(modelUser)
	return userSt, nil
}


func (h *UserServiceHandler) ModelUser2StructUset(modelUser *datamodels.User) *user.UserStruct {
	return &user.UserStruct{
		int64(modelUser.Id),
		modelUser.UserName,
		modelUser.NickName,
		int32(modelUser.State),
		modelUser.CreateAt,
		modelUser.UpdateAt,
	}
}

//   ./rpc_user_server -ip 127.0.0.1 -port 9090
func main() {
	flagIp := flag.String("ip", "", "rpc server ip")
	//指定监听端口
	flagPort := flag.Int("port", 0, "rpc server port")
	flag.Parse()

	if len(*flagIp) == 0 {
		log.Fatalln("flag ip error")
	}
	if *flagPort <= 0 {
		log.Fatalln("flag port error")
	}

	//本机的指定端口
	localIp := *flagIp
	localPort := *flagPort
	addr := fmt.Sprintf("127.0.0.1:%d", localPort)
	log.Println("listen: ", addr)

	//1. protocolFactory
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	//2. transportFactory
	transportFactory := thrift.NewTTransportFactory()

	//3. socket
	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		log.Fatalln("rpc user new socket fail", err)
	}

	//4. xxxHandler
	handler, err := NewUserServiceHandler()
	if err != nil {
		log.Fatalln("new userServiceHandler fail: ", err)
	}

	//5. xxxProcessor
	processor := user.NewUserServiceProcessor(handler)

	//6. NewTSimpleServer4
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	//这里直接知道服务名
	serviceConfig, err := handler.ConsulClient.Config.GetServiceConfigByName(handler.ConsulClient.Config.GetRpcUserServiceName())
	if err != nil {
		log.Fatalln("get service config fail: ", err)
	}
	log.Println("serviceConfig: ", serviceConfig)

	err = handler.ConsulClient.RegisterServer(serviceConfig.Name, serviceConfig.Tags, localIp, localPort)
	if err != nil {
		log.Fatalln("rpc user register service fail ", err)
	}

	server.Serve()
}








