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
	MysqlPool *common.MysqlPool
}

func NewUserServiceHandler(mysqlPool *common.MysqlPool) (userServiceHandler *UserServiceHandler, err error) {
	return &UserServiceHandler{mysqlPool}, nil
}


func (h *UserServiceHandler) Reg(ctx context.Context, userName, nickName, password string) (*user.UserStruct, error) {
	userService := services.NewUserService(repositories.NewUserRepository(h.MysqlPool))

	modelUser := &datamodels.User{
		UserName: userName,
		NickName: nickName,
		Password: password,
	}

	err := userService.InsertUser(modelUser)
	if err != nil {
		return &user.UserStruct{}, err
	}

	modelUser, err  = userService.GetUserByName(userName)
	if err != nil {
		return &user.UserStruct{}, err
	}

	userSt := h.ModelUser2StructUset(modelUser)
	return userSt, nil
}


func (h *UserServiceHandler) Login(ctx context.Context, userName , password string) (*user.UserStruct, error) {
	userService := services.NewUserService(repositories.NewUserRepository(h.MysqlPool))

	modelUser, isOk := userService.IsPwdSuccess(userName, password)
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

//   ./rpc_user_server -ip 127.0.0.1 -port 9001
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
	addr := fmt.Sprintf("%s:%d", localIp, localPort)
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

	config, err := common.NewConfigConsul()
	freeCache := common.NewFreeCacheClient(5)
	consulClient, err := common.NewConsulClient(config, freeCache)

	mysqlPool, err := common.NewMysqlPool(consulClient)
	if err != nil {

	}

	//4. xxxHandler
	handler, err := NewUserServiceHandler(mysqlPool)
	if err != nil {
		log.Fatalln("new userServiceHandler fail: ", err)
	}

	//5. xxxProcessor
	processor := user.NewUserServiceProcessor(handler)

	//6. NewTSimpleServer4
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	server.Serve()
}








