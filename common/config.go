package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
)

func init() {
	//开发环境
	if runtime.GOOS == "windows" {
		os.Setenv("CONF_DIR", "D:/Code/go/src/miaosha-demo/conf")
		os.Setenv("LOG_DIR", "D:/Code/go/src/miaosha-demo")
	} else {
		os.Setenv("LOG_DIR", "/var/log/miaosha-demo")

		out, err := exec.Command("/bin/bash", "-c", "ip addr show | grep 192.168.0.73").Output()
		if err == nil && len(out) > 0 {
			//remote
			os.Setenv("CONF_DIR", "/opt/code/miaosha-demo/conf")
		} else {
			//dev
			os.Setenv("CONF_DIR", "/opt/code/go/src/miaosha-demo/conf")
		}
	}
}


//consul的配置文件
func GetConfigFileConsul() (file string) {
	return os.Getenv("CONF_DIR") + "/consul.json"
}


func GetLogFile() (file string) {
	return os.Getenv("LOG_DIR") + "/miaosha.log"
}


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
	fileName := GetConfigFileConsul()
	consulByte, err := ioutil.ReadFile(fileName)
	if err != nil {
		ZapError("conf read fail", err)
		return nil, err
	}

	configConsul = &ConfigConsul{}
	err = json.Unmarshal(consulByte, configConsul)

	if err != nil {
		ZapError("conf consul unmarshal fail", err)
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

