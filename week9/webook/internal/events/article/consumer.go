package article

import (
	"context"
	"geekgo/week9/webook/internal/repository"
	"geekgo/week9/webook/pkgs/logger"
	"geekgo/week9/webook/pkgs/saramax/metrics"
	"github.com/IBM/sarama"
	"time"
)

// consumer调用的时候 比较复杂 sarama要初始化consumer组 还要setup consume等方法
// 这里在pkg里自己封装一个saramax consumer_handler 调用saramax里面的处理json解码的部分 传入consume函数

type Consumer interface {
	Start() error
}

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository // 消费的业务就是阅读数加1 这里调用repo阅读数+1
	l      logger.LoggerV1
}

func (i *InteractiveReadEventConsumer) Start() error {
	// 创建消费者组
	cg, err := sarama.NewConsumerGroupFromClient("interactive",
		i.client)
	if err != nil {
		return err
	}
	// 进行消费
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			//saramax.NewHandler[ReadEvent](i.l, i.Consume)) // saramax 封装了ConsumerGroupHandler的实现
			metrics.NewPrometheusKafkaConsumerHandler[ReadEvent](i.l, i.Consume))
		if err != nil {
			// 记录日志
			//r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", t.Aid)
}

func NewInteractiveReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{client: client, repo: repo, l: l}
}
