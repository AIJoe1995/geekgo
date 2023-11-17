package metrics

// 增加gin middleware prometheus 统计metrics 响应时间等
// 状态码可以一起在这里统计吗？ respbody 反序列化能拿到Result code吗

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

func NewMiddlewareBuilder(namespace string, subsystem string, name string, help string, instanceID string) *MiddlewareBuilder {
	return &MiddlewareBuilder{Namespace: namespace, Subsystem: subsystem, Name: name, Help: help, InstanceID: instanceID}
}

func (m *MiddlewareBuilder) Build() gin.HandlerFunc {
	labels := []string{"method", "pattern", "status"}
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_resp_time",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		}}, labels)

	prometheus.MustRegister(summary)
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_active_req",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		start := time.Now()
		gauge.Inc()

		defer func() {
			gauge.Dec()
			pattern := ctx.FullPath()
			if pattern == "" {
				pattern = "unknown"
			}
			summary.WithLabelValues(ctx.Request.Method,
				pattern,
				strconv.Itoa(ctx.Writer.Status())).Observe(float64(time.Since(start).Milliseconds()))
		}()

		ctx.Next()
	}
}
