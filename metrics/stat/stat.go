package stat

import (
	"github.com/iWuxc/go-wit/metrics/prometheus"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
)

var (
	// APPBootSeconds 服务启动时间
	APPBootSeconds = prometheus.NewGaugeFrom(
		stdPrometheus.GaugeOpts{
			Name: "app_boot_seconds",
			Help: "serving start info. ",
		},
		nil,
	)

	// APPBootInfo 服务启动信息
	APPBootInfo = prometheus.NewGaugeFrom(
		stdPrometheus.GaugeOpts{
			Name: "app_boot_info",
			Help: "serving start info. ",
		},
		[]string{"app_port", "app_version", "app_pid", "app_name", "kit_version"},
	)

	// APPErrorCount 服务错误统计
	APPErrorCount = prometheus.NewCounterFrom(
		stdPrometheus.CounterOpts{
			Name: "app_error_total",
			Help: "total number of errors in the server.",
		},
		[]string{"location", "level", "info"},
	)

	// APPRequestTotalCount 服务请求统计
	APPRequestTotalCount = prometheus.NewCounterFrom(
		stdPrometheus.CounterOpts{
			Name: "app_request_total",
			Help: "the server received request num with every uri.",
		},
		[]string{"uri", "code"},
	)

	// APPRequestHistogram 服务请求响应时长
	APPRequestHistogram = prometheus.NewHistogramFrom(
		stdPrometheus.HistogramOpts{
			Name:    "app_request_durations",
			Help:    "the time server took to handle the request.",
			Buckets: []float64{0.2, 0.5, 1, 3, 8, 20},
		},
		[]string{"uri"},
	)
)
