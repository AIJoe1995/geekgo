package events

import (
	"context"
	"geekgo/week19/follow/repository"
	"geekgo/week19/follow/repository/dao"
	"geekgo/week19/pkg/canalx"
	"geekgo/week19/pkg/saramax"
	"github.com/IBM/sarama"
	"time"
)

type MYSQLBinlogConsumer struct {
	client sarama.Client
	repo   repository.FollowRepository
}

func NewMYSQLBinlogConsumer(client sarama.Client, repo repository.FollowRepository) *MYSQLBinlogConsumer {
	return &MYSQLBinlogConsumer{client: client, repo: repo}
}

func (r *MYSQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("follow_statics",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"webook_binlog"},
			saramax.NewHandler[canalx.Message[dao.FollowRelation]](r.Consume))
		if err != nil {

		}
	}()
	return err
}

func (r *MYSQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage,
	val canalx.Message[dao.FollowRelation]) error {
	if val.Table != "follow_relations" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, data := range val.Data {
		var err error
		switch data.Status {
		case dao.FollowRelationStatusActive:
			err = r.repo.SetFollowerStatics(ctx, data.Follower, 1)
			err = r.repo.SetFolloweeStatics(ctx, data.Followee, 1)
		case dao.FollowRelationStatusInactive:
			err = r.repo.SetFollowerStatics(ctx, data.Follower, -1)
			err = r.repo.SetFolloweeStatics(ctx, data.Followee, -1)

		}
		if err != nil {
			return err
		}
	}
	return nil
}
