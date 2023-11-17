package metrics

import (
	"encoding/json"
	"geekgo/week9/webook/pkgs/logger"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var counter *prometheus.CounterVec
var summary *prometheus.SummaryVec

func InitCounter(opt prometheus.CounterOpts) {
	counter = prometheus.NewCounterVec(opt,
		[]string{"topic"})
	prometheus.MustRegister(counter)
	// 你可以考虑使用 code, method, 命中路由，HTTP 状态码
}

func InitSummary(opt prometheus.SummaryOpts) {
	summary = prometheus.NewSummaryVec(opt, []string{"topic"})
	prometheus.MustRegister(summary)
}

type PrometheusKafkaConsumerHandler[T any] struct {
	count   *prometheus.CounterVec
	summary *prometheus.SummaryVec
	l       logger.LoggerV1
	fn      func(msg *sarama.ConsumerMessage, t T) error
}

func (h *PrometheusKafkaConsumerHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {

		startTime := time.Now()

		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列化消息失败",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset))
			continue
		}

		//	consume fn 重试
		for i := 0; i < 3; i++ {
			err = h.fn(msg, t)
			if err == nil {
				break
			}
			h.count.WithLabelValues(msg.Topic).Inc()
			h.l.Error("处理消息失败",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset))

		}
		if err != nil {
			h.l.Error("处理消息失败-重试次数上限",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int64("partition", int64(msg.Partition)),
				logger.Int64("offset", msg.Offset))
		} else {
			session.MarkMessage(msg, "")
		}
		h.summary.WithLabelValues(msg.Topic).Observe(float64(time.Since(startTime).Milliseconds()))
	}
	return nil

}

func (h *PrometheusKafkaConsumerHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil

}

func (h *PrometheusKafkaConsumerHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil

}

func NewPrometheusKafkaConsumerHandler[T any](l logger.LoggerV1, fn func(msg *sarama.ConsumerMessage, t T) error) *PrometheusKafkaConsumerHandler[T] {
	return &PrometheusKafkaConsumerHandler[T]{
		l:       l,
		fn:      fn,
		count:   counter,
		summary: summary,
	}
}
