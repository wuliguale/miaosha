package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
	"miaosha-demo/services"
	"time"
)

func main() {
	flagRed := flag.Int("read", 1, "is read")
	flagNum := flag.Int("num", 10, "num")
	flag.Parse()

	config, err := common.NewConfigConsul()
	if err != nil {
		common.ZapError("new config fail", err)
		return
	}

	freecache := common.NewFreeCacheClient(20)
	consul, err := common.NewConsulClient(config, freecache)
	if err != nil {
		common.ZapError("new consul fail", err)
		return
	}

	mysqlPool, err := common.NewMysqlPoolProduct(consul)
	if err != nil {
		common.ZapError("new mysql pool fail", err)
		return
	}

	if *flagRed == 1 {
		read(*flagNum, mysqlPool)
	} else {
		write(*flagNum, mysqlPool)
	}
}


func read(num int, mysqlPool *common.MysqlPool) {
	repo := repositories.NewUserRepository(mysqlPool)
	service := services.NewUserService(repo)

	timeStart := time.Now()
	for i := 0; i < num; i++ {
		user, err := service.GetUserByName(time.Now().String())
		if err != nil {
			common.ZapError("get user from db fail", err)
			continue
		}

		zap.L().Info(user.NickName)
	}
	timeEnd := time.Now()

	timeTotal := timeEnd.Sub(timeStart).Milliseconds()
	timeAvg := timeTotal / int64(num)

	fmt.Println(fmt.Sprintf("mysql read: %d, time total: %d ms, time avg: %d ms", num, timeTotal, timeAvg))
}


func write(num int, mysqlPool *common.MysqlPool) {
	repo := repositories.NewOrderRepository(mysqlPool)
	service := services.NewOrderService(repo)

	timeStart := time.Now()
	for i := 0; i < num; i++ {
		uid := i
		pid := i * 2

		order := &datamodels.Order{}
		order.Uid = uint32(uid)
		order.Pid = uint64(pid)
		order.State = datamodels.OrderWait
		order.CreateAt = time.Now().Unix()

		err := service.InsertIgnoreOrder(order)
		if err != nil {
			common.ZapError("write order to db fail", err)
		}
	}
	timeEnd := time.Now()
	timeTotal := timeEnd.Sub(timeStart).Milliseconds()
	timeAvg := timeTotal / int64(num)

	fmt.Println(fmt.Sprintf("mysql write: %d, time total: %d ms, time avg: %d ms", num, timeTotal, timeAvg))
}


