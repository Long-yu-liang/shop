package main

import (
	"flag"
	"fmt"
	"net"
	"shop_servs/user_srv/global"
	"shop_servs/user_srv/handler"
	"shop_servs/user_srv/initialize"
	proto "shop_servs/user_srv/proto/user"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	//解析用户参数
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")

	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	flag.Parse()
	zap.S().Info("ip:", *IP)
	zap.S().Info("port:", *Port)
	//服务流程
	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	//监听
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}
	//注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//服务注册
	// 1. Consul 客户端配置
	cfg := api.DefaultConfig()
	//consul集群地址
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	// 2. 健康检查配置
	check := &api.AgentServiceCheck{ //健康检查地址，告诉 Consul 如何检查我的健康状态
		GRPC:                           "192.168.0.3:50051",
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "15s",
	}
	// 3. 服务注册配置
	registration := &api.AgentServiceRegistration{
		Name:    global.ServerConfig.Name,
		ID:      global.ServerConfig.Name,
		Port:    *Port,
		Tags:    []string{"lyl", "bobby", "user", "srv"},
		Address: "192.168.0.3", //服务注册地址，告诉其他服务如何访问我
		Check:   check,
	}
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	//启动服务
	err = server.Serve(lis)
	if err != nil {
		panic("failed to start grpc:" + err.Error())
	}

}
