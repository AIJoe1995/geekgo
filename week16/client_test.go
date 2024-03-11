package main

import (
	"context"
	userGRPC "geekgo/week16/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

type EtcdClientTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdClientTestSuite) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *EtcdClientTestSuite) TestCustomRoundRobinClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	svcCfg := `
{
    "loadBalancingConfig": [
        {
            "custom_wrr": {}
        }
    ]
}
`
	cc, err := grpc.Dial("etcs:///service/user", grpc.WithResolvers(bd),
		grpc.WithDefaultServiceConfig(svcCfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := userGRPC.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		resp, err := client.GetById(ctx, &userGRPC.GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func TestEtcdClient(t *testing.T) {
	suite.Run(t, new(EtcdClientTestSuite))
}
