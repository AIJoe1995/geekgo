package ioc

import (
	"geekgo/week19/follow/events"
	"geekgo/week19/pkg/saramax"
	"github.com/IBM/sarama"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	cfg.Addrs = []string{"localhost:9092"}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewConsumers(follow *events.MYSQLBinlogConsumer) []saramax.Consumer {
	return []saramax.Consumer{
		follow,
	}
}
