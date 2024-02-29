package fixer

import (
	"context"
	"encoding/json"
	"errors"
	"geekgo/week13/migrator"
	"geekgo/week13/migrator/events"
	"geekgo/week13/migrator/fixer"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"time"
)

// 消费InconsistentEvent
// 消费的逻辑在fixer包中fix.go的Fix函数

// Consumer 因为需要根据迁移过程 双写时 先以源表为准 后以目标表为准 所以修复的时候 src dst交换位置
type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](
	client sarama.Client,
	topic string,
	src *gorm.DB,
	dst *gorm.DB) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{
		client:   client,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
	}, nil
}

func (r *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", r.client)
	if err != nil {
		return err
	}
	go func() {
		err1 := cg.Consume(context.Background(),
			[]string{r.topic},
			&Handler[events.InconsistentEvent]{fn: r.Consume})
		if err1 != nil {
			// 退出了消费循环 记录日志
		}
	}()
	return err // 这个err
}

// 消费 fix数据 需要区分 源表为准 目标表为准
func (r *Consumer[T]) Consume(msg *sarama.ConsumerMessage, t events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch t.Direction {
	case "SRC":
		return r.srcFirst.Fix(ctx, t.ID, t.Type)
	case "DST":
		return r.dstFirst.Fix(ctx, t.ID, t.Type)
	}
	return errors.New("未知的校验方向")
}

type Handler[T any] struct {
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			// 消息格式都不对，没啥好处理的
			// 但是也不能直接返回，在线上的时候要继续处理下去

			// 不中断，继续下一个
			session.MarkMessage(msg, "")
			continue
		}
		err = h.fn(msg, t)
		if err != nil {
			//记录日志
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
