package ioc

import (
	events "geekgo/week9/webook/internal/events/article"
	"github.com/IBM/sarama"
)

func InitKafka() sarama.Client {

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true

	client, err := sarama.NewClient([]string{"localhost:9094"}, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(c1 *events.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
