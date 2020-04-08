package common

import (
	"github.com/go-redis/redis/v7"
	"strconv"
)

func NewRedisClusterClient(consul *ConsulClient) (*redis.ClusterClient, error) {
	serviceName := consul.Config.GetRedisServiceName()
	serviceInfoList, err := consul.GetServiceListByName(serviceName)
	if err != nil {
		return nil, err
	}

	var addrList []string
	for _,serviceInfo := range serviceInfoList.List {
		addrList = append(addrList, serviceInfo.Host + ":" + strconv.Itoa(serviceInfo.Port))
	}

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs : addrList,
		Password: "310900",
	}), nil
}
