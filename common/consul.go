package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"sync"
	"time"
)

//一个服务的详情
type ConsulServiceInfo struct {
	Id string
	Name string
	Host string
	Port int
	Tags []string
}


//同一个服务的列表
type ConsulServiceInfoList struct {
	Name string
	//服务列表
	List []*ConsulServiceInfo
	//下一个服务的位置
	Next int
	sync.RWMutex
}


func (list *ConsulServiceInfoList) Add(serviceInfo *ConsulServiceInfo) {
	list.Lock()
	list.List = append(list.List, serviceInfo)
	list.Unlock()
}


func (list *ConsulServiceInfoList) Remove(index int) {
	if index < 0 {
		return
	}

	list.Lock()
	list.List = append(list.List[:index], list.List[index+1:]...)

	if index == len(list.List){
		//删除最后一个
		list.Next = 0
	} else if index < list.Next {
		//删除next之前的，next前移一位
		list.Next--
	}

	list.Unlock()
}

func (list *ConsulServiceInfoList) Clear() {
	list.Lock()
	list.List = list.List[:0]
	list.Next = 0
	list.Unlock()
}


func (list *ConsulServiceInfoList) IsEmpty() (isEmpty bool) {
	list.RLock()
	if len(list.List) == 0 {
		isEmpty = true
	} else {
		isEmpty = false
	}
	list.RUnlock()

	return isEmpty
}

//取下一个可用元素
func (list *ConsulServiceInfoList) GetNext() (serviceInfo *ConsulServiceInfo, err error) {
	if list.IsEmpty() {
		return nil, errors.New("service list empty")
	}

	list.Lock()
	serviceInfo = list.List[list.Next]
	list.Next = (list.Next + 1) % len(list.List)
	list.Unlock()

	return serviceInfo, nil
}


//consul客户端
type ConsulClient struct {
	*api.Client
	Config *ConfigConsul
	Cache *FreeCacheClient
	//serviceName:ServiceInfoList
	ChanList map[string]chan *ConsulServiceInfoList
}

func NewConsulClient(configConsul *ConfigConsul, freeCache *FreeCacheClient) (consul *ConsulClient, err error) {
	config := api.DefaultConfig()

	//从配置文件取，agent的负载均衡地址
	config.Address = fmt.Sprintf("%s:%d", configConsul.Host, configConsul.Port)
	fmt.Println(config.Address)

	client, err := api.NewClient(config)

	if err != nil {
		return nil, err
	} else {
		ChanList := map[string]chan *ConsulServiceInfoList{}
		consul = &ConsulClient{client, configConsul, freeCache, ChanList}

		//一直watch consul上的service
		serviceNameList := configConsul.GetServiceNameList()
		for _, serviceName := range serviceNameList {
			//chan只保留最新的serviceInfoList
			consul.ChanList[serviceName] = make(chan *ConsulServiceInfoList, 1)

			go consul.WatchServiceByName(serviceName)
		}

		return consul, nil
	}
}


//取一个service
func (client *ConsulClient) GetServiceByName(serviceName string) (serviceInfo *ConsulServiceInfo, err error) {
	serviceInfoList, err := client.GetServiceListByName(serviceName)
	if err != nil {
		log.Fatalln("get service info list fail: ", err)
	}

	return serviceInfoList.GetNext()
}


func (client *ConsulClient) GetServiceListByName (serviceName string) (serviceInfoList *ConsulServiceInfoList, err error) {
	//先从cache取
	serviceConfig, err := client.Config.GetServiceConfigByName(serviceName)
	if err != nil {
		log.Fatalln("get service config fail: ", err)
	}

	cacheKey := serviceConfig.CacheKey
	cacheData, err := client.Cache.Get([]byte(cacheKey))
	if err != nil && !client.Cache.IsNotFound(err) {
		log.Fatalln("get service list from cache fail ", err)
	}

	//缓存没有就从线上consul取
	if client.Cache.IsNotFound(err) {
		log.Println("get cache not found")

		serviceInfoList, err = client.FetchServiceInfoListByName(serviceName)
		if err != nil {
			log.Fatalln("fetch service list from consul fail", err)
		}

		//保存到cache
		err = client.CacheServiceInfoList(serviceInfoList)
		if err != nil {
			log.Fatalln("cache service info list fail ", err)
		}
		return serviceInfoList, nil
	}

	log.Println("get cache succ")

	//返回cache取出的数据
	serviceInfoList = &ConsulServiceInfoList{}
	err = json.Unmarshal(cacheData, serviceInfoList)
	if err != nil {
		log.Fatalln("get service info list from cache unmarshal fail ", err)
	}

	return serviceInfoList, err
	//TODO 从硬盘中存取服务信息，防止cache,consul失效的情况
}


func (client *ConsulClient) FetchServiceInfoListByName(serviceName string) (serviceInfoList *ConsulServiceInfoList, err error) {
	serviceList, _, err := client.Health().Service(serviceName, "", true, &api.QueryOptions{})

	if err != nil {
		return nil, err
	}

	serviceInfoList = client.FormatApiServiceList2ServiceInfoList(serviceName, serviceList)
	return serviceInfoList, nil
}



func (client *ConsulClient) WatchServiceByName(serviceName string) {
	//waitIndex从0开始，类似版本号
	var lastIndex uint64 = 0

	serviceConfig, err := client.Config.GetServiceConfigByName(serviceName)
	if err != nil {
		log.Println("get service config fail: ", err)
	}

	//从配置文件取等待时间和休息时间
	waitSeconds := time.Duration(serviceConfig.WaitSeconds) * time.Second
	sleepSeconds := time.Duration(serviceConfig.SleepSeconds) * time.Second

	for {
		//服务有变化（WaitIndex变化），或持续等待WaitTime结束，返回最新数据（没变化也返回数据）
		// 没有服务不报错，没有服务也有lastIndex
		serviceList, metaInfo, err := client.Health().Service(serviceName, "", true, &api.QueryOptions{
			WaitIndex:lastIndex,
			WaitTime:waitSeconds,
		})

		if err != nil {
			log.Println("service watch error: ", err)
		}

		//log.Println(serviceName, lastIndex, metaInfo.LastIndex, len(serviceList), err)

		//数据有变化才写入，避免频繁写入，如果cache过期可以在get时加入，不需要watch时一直写
		if lastIndex != metaInfo.LastIndex{
			//有数据
			if len(serviceList) >= 0 {
				serviceInfoList := client.FormatApiServiceList2ServiceInfoList(serviceName, serviceList)

				//将服务的最新数据写入服务的chan，以便服务使用这更新
				client.SendServiceInfoList2Chan(serviceName, serviceInfoList)

				err = client.CacheServiceInfoList(serviceInfoList)
				if err != nil {
					log.Println("cache serviceInfoList fail ", err)
				}

				log.Println("watch cache serviceInfoList", serviceInfoList)
			}
		}

		lastIndex = metaInfo.LastIndex
		if sleepSeconds > 0 {
			time.Sleep(sleepSeconds)
		}
	}
}


func (client *ConsulClient) SendServiceInfoList2Chan(serviceName string, serviceInfoList *ConsulServiceInfoList) (err error) {
	err = errors.New("send serviceInfoList to chan fail")

	//clear chan
	select {
	case <- client.ChanList[serviceName]:
	default:
	}

	log.Println("consul send serviceInfoList to chan2", serviceName)

	select {
	case client.ChanList[serviceName] <- serviceInfoList:
		fmt.Println("consul send serviceInfoList to chan succ")
		err = nil
	case <- time.After(1 * time.Second):
		fmt.Println("consul send serviceInfoList to chan timeout")
	}

	return err
}



//将从consul获取的service列表转换为ServiceInfoList
func (client *ConsulClient) FormatApiServiceList2ServiceInfoList (serviceName string, serviceSlice []*api.ServiceEntry) (serviceInfoList *ConsulServiceInfoList){
	serviceInfoList = &ConsulServiceInfoList{Name:serviceName}

	for _, service := range serviceSlice {
		serviceInfo := &ConsulServiceInfo{}

		serviceInfo.Id = service.Service.ID
		serviceInfo.Name = service.Service.Service
		serviceInfo.Host = service.Service.Address
		serviceInfo.Port = service.Service.Port
		serviceInfo.Tags = service.Service.Tags

		serviceInfoList.Add(serviceInfo)
	}

	return serviceInfoList
}


//存到cache
func (client *ConsulClient) CacheServiceInfoList (serviceInfoList *ConsulServiceInfoList) (err error){
	cacheDataByte, err := json.Marshal(*serviceInfoList)
	if err != nil {
		return err
	}

	serviceConfig, err := client.Config.GetServiceConfigByName(serviceInfoList.Name)
	if err != nil {
		log.Fatalln("get service config fail: ", err)
	}

	//get cache key from config
	cacheKey := serviceConfig.CacheKey
	//get cache seconds from config
	cacheSeconds := serviceConfig.CacheSeconds

	err = client.Cache.Set([]byte(cacheKey), cacheDataByte, cacheSeconds)
	return err
}


func (client *ConsulClient) GetRegisterServiceId(serviceName, localIp string, localPort int) string {
	return fmt.Sprintf("service-%s-%s:%d", serviceName, localIp, localPort)
}


func (client *ConsulClient) GetRegisterCheckId(serviceName, localIp string, localPort int) string {
	return fmt.Sprintf("check-%s-%s:%d", serviceName, localIp, localPort)
}


//注册服务
func (client *ConsulClient) RegisterServer(serviceName string, serviceTags []string, localIp string, localPort int) error{
	//localIp := GetLocalIP()

	//服务定义
	registration := &api.AgentServiceRegistration{}
	//service id不同，service name相同，使用service name查询时可以得到多个service
	registration.ID = client.GetRegisterServiceId(serviceName, localIp, localPort)
	registration.Name = serviceName
	registration.Tags = serviceTags
	registration.Address = localIp
	registration.Port = localPort

	serviceConfig, err := client.Config.GetServiceConfigByName(serviceName)
	if err != nil {
		log.Fatalln("get service config fail: ", err)
	}

	//健康检查
	checkId := client.GetRegisterCheckId(serviceName, localIp, localPort)
	checkOutput := serviceConfig.CheckOutput + checkId
	ttl := serviceConfig.CheckTtlSeconds
	checkRealSeconds := serviceConfig.CheckRealSeconds

	registration.Check = &api.AgentServiceCheck{
		CheckID:checkId,
		TTL:fmt.Sprintf("%ds", ttl),
	}

	//服务注册，重复则覆盖
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalln("rpc user register server fail ", err)
	}

	//使用ttl刷新服务状态
	go func() {
		for range time.Tick(time.Duration(checkRealSeconds) * time.Second){
			err = client.Agent().UpdateTTL(checkId, checkOutput, api.HealthPassing)
		}
	}()

	return nil
}

//注销服务
func (client *ConsulClient) DeregisterService (serviceName, localIp string, localPort int) error {
	serviceId := client.GetRegisterServiceId(serviceName, localIp, localPort)
	err := client.Agent().ServiceDeregister(serviceId)

	if err != nil {
		log.Fatalln("rpc user deregister server  fail", err)
	}

	return nil
}

