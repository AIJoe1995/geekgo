package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

// 初始化 对不同类型的Event[T] 调用不同的处理函数fn
type Handler[T any] struct {
	fn func(message *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{fn: fn}
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
			// 记日志
			// 不中断，继续下一个
			session.MarkMessage(msg, "")
			continue
		}
		// 使用Handler中的fn处理消息
		err = h.fn(msg, t)
		session.MarkMessage(msg, "")

	}
	return nil
}
