package main

import (
	"context"
	userGRPC "geekgo/week16/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}

func (s *EtcdTestSuite) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *EtcdTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 20)
	}()
	go func() {
		s.startServer(":8091", 10)
	}()
}

func GetOutBoundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (s *EtcdTestSuite) startServer(addr string, weight int) {
	l, err := net.Listen("tcp", addr)
	require.NoError(s.T(), err)

	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(s.T(), err)
	addr = GetOutBoundIP() + addr
	key := "service/user/" + addr

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	var ttl int64 = 30
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(s.T(), err)
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]interface{}{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(s.T(), err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err1)

	}()

	server := grpc.NewServer()

	userGRPC.RegisterUserServiceServer(server, &userGRPC.Server{
		Name: addr,
	})

	err = server.Serve(l)
	s.T().Log(err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	s.client.Close()
	server.GracefulStop()
}
