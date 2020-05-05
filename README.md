# 秒杀demo

## 一、功能与实现
使用Go框架Iris开发，基于Docker搭建集群

#### 秒杀流程
1. 将商品信息放入Redis
2. 检查秒杀是否已结束
3. 检查用户是否已参加过，incr结果大于1表示已参加过
4. 检查库存，decr结果大于等于0表示秒杀成功
5. 发送商品和用户信息到RabbitMQ
6. 从RabbitMQ中取出商品和用户信息，生成订单


#### Kong：负载均衡，jwt鉴权，流量控制
1. Kong集群，通过连接相同的PostgreSQL共享配置，可水平扩展，修改不用重启
2. 通过DNS或LVS将流量分配到不同的Kong实例，Kong通过upstream将流量分配到不同的后端服务器上
3. 启用jwt插件，检查秒杀请求是否有效
4. 启用rate-limiting插件，使用Redis实现按ip的流量控制


#### Consul：服务发现
1. Consul集群，3个server agent，每个服务通过同机器的client agent注册和检查服务
2. LVS+Keepalived实现集群内Consul server agent的负载均衡

#### ProxySQL：MySQL中间件
1. 集群，使用相同配置，可水平扩展
2. 读写分离，按权重分配流量
3. 监控MySQL主从切换，并自动调整读写
4. 所有实例注册到Consul


#### MGR：MySQL高可用
1个primary member写，2个非primary member读，自动切换


#### RabbitMQ：消息队列
1. 集群，3个disk实例，3个ram实例，可水平扩展ram实例
2. 镜像队列，保证消息不丢失
3. 备份队列和死信队列
4. LVS+Keepalived实现集群内ram实例的负载均衡，VIP注册到Consul



#### Redis Cluster：Redis集群
1. 3个master实例， 3个slave实例，master读写，slave只备份不读写
2. 只将master实例注册到Consul

#### Thrift：RPC
1. 用户服务实现注册和登录功能，可启动多实例组成集群
2. 所有用户服务实例注册到Consul


#### 连接池
1. 单例使用
2. 从consul取服务列表，按服务列表创建连接，chan保存创建连接，使用连接时get，用完时put
3. get时可检查连接是否可用和是否超过空闲时间
4. consul上服务列表变化时，通过chan通知连接池，连接池使用新服务列表重新生成连接并放入连接池


------------

## 二、目录结构
conf/，配置文件目录

common/，通用功能目录<br>
　　config，配置文件读取等操作<br>
　　**consul**，服务发现和服务监控（watch）<br>
　　freecache，freeCache相关功能，可设置缓存过期时间<br>
　　jwt，kong所使用的jwt生成和解析<br>
　　**mysql**，mysql连接池<br>
　　**pool**，通用连接池<br>
　　**rabbitmq**，rabbitmq连接池<br>
　　**redis**，redis的smart客户端<br>

fronted/，前台功能<br>
　　web/<br>
  　　　controllers／<br>
　　　　　　**product**,秒杀接口，和商品操作相关功能<br>
　　　　　　**user**，用户登录和注册页面（thrift rpc调用用户服务）<br>
　　　　views／<br>
　　**main**，前台main<br>
　　**mq_receive**，订单服务，从rabbitmq消费数据生成订单<br>
　　**rpc_user_server**，用户服务，通过thrift rpc被调用<br>

backend/，后台功能<br>

rpc/，rpc相关<br>
　　gen-go/，thrift生成的服务端代码<br>
　　user.thrift，定义的thrift服务<br>
　　**user-client**，thrift客户端访问user服务的代码<br>

datamodels/<br>
repositories/<br>
services/<br>


------------


## 三、架构图
[miaosha-demo图](https://note.youdao.com/ynoteshare1/index.html?id=33026920739ca35a73440fffbf326869)








