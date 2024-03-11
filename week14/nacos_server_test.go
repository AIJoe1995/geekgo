package main

import (
	userGRPC "geekgo/week14/grpc"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"net"
	"testing"
)

type NacosServerTestSuite struct {
	suite.Suite
	client naming_client.INamingClient
}

func (s *NacosServerTestSuite) SetupSuite() {
	//create clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}
	// At least one ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}
	cli, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	require.NoError(s.T(), err)
	s.client = cli
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (s *NacosServerTestSuite) TestServer() {
	l, err := net.Listen("tcp", ":8091")
	require.NoError(s.T(), err)
	server := grpc.NewServer()
	userGRPC.RegisterUserServiceServer(server, &userGRPC.Server{})
	ok, err := s.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          GetOutboundIP(),
		Port:        8091,
		ServiceName: "user",
		Enable:      true,
		Healthy:     true,
		Weight:      10,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), ok)
	err = server.Serve(l)
	s.T().Log(err)
}

func TestNacosServer(t *testing.T) {
	suite.Run(t, new(NacosServerTestSuite))
}
