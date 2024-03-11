package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

type Handler[T any] struct {
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{

		fn: fn,
	}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 可以考虑在这个封装里面提供统一的重试机制
func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
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

		}
		session.MarkMessage(msg, "")
	}
	return nil
}