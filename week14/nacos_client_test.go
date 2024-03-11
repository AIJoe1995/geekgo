package main

import (
	"context"
	userGRPC "geekgo/week14/grpc"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

type NacosClientTestSuite struct {
	suite.Suite
	client naming_client.INamingClient
}

func (s *NacosClientTestSuite) SetupSuite() {
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

func (s *NacosClientTestSuite) TestClient() {
	rb := &nacosResolverBuilder{
		client: s.client,
	}
	cc, err := grpc.Dial("nacos:///user",
		grpc.WithResolvers(rb),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
	client := userGRPC.NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &userGRPC.GetByIdRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}

func TestNacosClient(t *testing.T) {
	suite.Run(t, new(NacosClientTestSuite))
}
