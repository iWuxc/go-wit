package prometheus

import (
	"github.com/iWuxc/go-wit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var _ metrics.Counter = (*counter)(nil)

type counter struct {
	cv  *prometheus.CounterVec
	lvs LabelValues
}

// NewCounterFrom constructs and registers a Prometheus CounterVec,
// and returns a usable Counter object.
func NewCounterFrom(opts prometheus.CounterOpts, labelNames []string) metrics.Counter {
	cv := prometheus.NewCounterVec(opts, labelNames)
	prometheus.MustRegister(cv)
	return NewCounter(cv)
}

// NewCounter new a prometheus counter and returns Counter.
func NewCounter(cv *prometheus.CounterVec) metrics.Counter {
	return &counter{
		cv: cv,
	}
}

func (c *counter) With(lvs ...string) metrics.Counter {
	return &counter{
		cv:  c.cv,
		lvs: c.lvs.With(lvs...),
	}
}

func (c *counter) Inc() {
	c.cv.With(makeLabels(c.lvs...)).Inc()
}

func (c *counter) Add(delta float64) {
	c.cv.With(makeLabels(c.lvs...)).Add(delta)
}
