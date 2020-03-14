package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
	"miaosha-demo/rpc/gen-go/user"
	"miaosha-demo/services"
)

//rpc UserService的服务端
type UserServiceHandler struct {
	UserService services.IUserService
}

func NewUserServiceHandler() *UserServiceHandler {
	db, err := common.NewMysqlConn()
	fmt.Println(err)

	userService := services.NewUserService(repositories.NewUserRepository(db))
	return &UserServiceHandler{userService}
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


func main() {
	addr := "127.0.0.1:9090"

	//1. protocolFactory
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	//2. transportFactory
	transportFactory := thrift.NewTTransportFactory()

	//3. socket
	transport, err := thrift.NewTServerSocket(addr)
	fmt.Println(err)

	//4. xxxHandler
	handler := NewUserServiceHandler()

	//5. xxxProcessor
	processor := user.NewUserServiceProcessor(handler)

	//6. NewTSimpleServer4
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	fmt.Println("Starting the simple server... on ", addr)

	//7. serve
	server.Serve()
}

