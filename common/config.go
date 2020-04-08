package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
)


type ConfigConsulService struct {
	Name string	`json:"name"`
	Tags []string `json:"tags"`
	WaitSeconds int `json:"wait_seconds"`
	SleepSeconds int `json:"sleep_seconds"`
	CheckTtlSeconds int `json:"check_ttl_seconds"`
	CheckRealSeconds int `json:"check_real_seconds"`
	CheckOutput string `json:"check_output"`
	CacheKey string `json:"cache_key"`
	CacheSeconds int `json:"cache_seconds"`
}

type ConfigConsul struct {
	Host string `json:"host"`
	Port int `json:"port"`
	Services map[string]ConfigConsulService `json:"services"`
}


func NewConfigConsul() (configConsul *ConfigConsul, err error){
	//TODO 配置文件统一存放
	mainPath, _ := os.Getwd()
	fileName := mainPath + "/../conf/consul.json"
	consulByte, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalln("conf read fail", err)
		return nil, err
	}

	configConsul = &ConfigConsul{}
	err = json.Unmarshal(consulByte, configConsul)

	if err != nil {
		log.Fatalln("conf consul unmarshal fail ", err)
		return nil, err
	}

	return configConsul, nil
}


func (configConsul *ConfigConsul) GetServiceNameList() (serviceNameList []string){
	for serviceName := range configConsul.Services {
		serviceNameList = append(serviceNameList, serviceName)
	}

	return serviceNameList
}


func (configConsul *ConfigConsul) GetServiceConfigByName(serviceName string) (serviceConfig ConfigConsulService, err error) {
	serviceConfig = ConfigConsulService{}
	serviceConfig, isOk := configConsul.Services[serviceName]

	if !isOk {
		return serviceConfig, errors.New("service not found")
	}

	return serviceConfig, nil
}


func (configConsul *ConfigConsul) GetRpcUserServiceName() string {
	return "miaosha-demo-user"
}


func (configConsul *ConfigConsul) GetRedisServiceName() string {
	return "miaosha-demo-redis"
}

