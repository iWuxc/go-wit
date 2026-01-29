package prometheus

import (
	"github.com/iWuxc/go-wit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var _ metrics.Observer = (*histogram)(nil)

type histogram struct {
	hv  *prometheus.HistogramVec
	lvs LabelValues
}

// NewHistogramFrom constructs and registers a Prometheus HistogramVec,
// and returns a usable Histogram object.
func NewHistogramFrom(opts prometheus.HistogramOpts, labelNames []string) metrics.Observer {
	hv := prometheus.NewHistogramVec(opts, labelNames)
	prometheus.MustRegister(hv)
	return NewHistogram(hv)
}

// NewHistogram new a prometheus histogram and returns Histogram.
func NewHistogram(hv *prometheus.HistogramVec) metrics.Observer {
	return &histogram{
		hv: hv,
	}
}

func (h *histogram) With(lvs ...string) metrics.Observer {
	return &histogram{
		hv:  h.hv,
		lvs: h.lvs.With(lvs...),
	}
}

func (h *histogram) Observe(value float64) {
	h.hv.With(makeLabels(h.lvs...)).Observe(value)
}
