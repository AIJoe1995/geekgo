package week19

import (
	"context"
	"geekgo/week19/follow/events"
	"geekgo/week19/follow/ioc"
	"geekgo/week19/follow/repository"
	"geekgo/week19/follow/repository/dao"
	"geekgo/week19/follow/service"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
)

type FollowTestSuite struct {
	suite.Suite
	db           *gorm.DB
	saramaClient sarama.Client
	svc          service.FollowRelationService
}

func TestFollow(t *testing.T) {
	suite.Run(t, new(FollowTestSuite))
}

func (s *FollowTestSuite) SetupSuite() {
	s.db = ioc.InitDB()
	s.saramaClient = ioc.InitKafka()

	d := dao.NewGORMFollowRelationDAO(s.db)
	repo := repository.NewFollowRelationRepository(d)
	svc := service.NewFollowRelationService(repo)
	s.svc = svc
	consumers := ioc.NewConsumers(events.NewMYSQLBinlogConsumer(s.saramaClient, repo))
	for _, c := range consumers {
		c.Start()
	}

}

func (s *FollowTestSuite) TestFollowStaticsCanal() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := s.svc.Follow(ctx, 1, 2)
	require.NoError(s.T(), err)
	stats1, err := s.svc.GetFollowStatics(ctx, 1)
	require.NoError(s.T(), err)
	stats2, err := s.svc.GetFollowStatics(ctx, 2)
	require.NoError(s.T(), err)
	s.T().Log(stats1)
	s.T().Log(stats2)

}
