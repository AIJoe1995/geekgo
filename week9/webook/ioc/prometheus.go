package ioc

import (
	"geekgo/week9/webook/pkgs/ginx"
	"geekgo/week9/webook/pkgs/saramax/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func InitPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func InitGinPrometheus() {
	opt := prometheus.CounterOpts{
		Namespace: "week9",
		Subsystem: "webook",
		Name:      "http" + "_req_count",
	}
	ginx.InitCounter(opt)
}

func InitKafkaPromethues() {
	CntOpt := prometheus.CounterOpts{
		Namespace: "week9",
		Subsystem: "webook",
		Name:      "kafka" + "_error_count",
	}

	sumOpt := prometheus.SummaryOpts{
		Namespace: "week9",
		Subsystem: "webook",
		Name:      "kafka" + "_consume_time",
		ConstLabels: map[string]string{
			"instance_id": "1",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}

	metrics.InitCounter(CntOpt)
	metrics.InitSummary(sumOpt)
}
