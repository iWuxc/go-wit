package prometheus

import (
	"github.com/iWuxc/go-wit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var _ metrics.Gauge = (*gauge)(nil)

type gauge struct {
	gv  *prometheus.GaugeVec
	lvs LabelValues
}

// NewGaugeFrom constructs and registers a Prometheus GaugeVec,
// and returns a usable Gauge object.
func NewGaugeFrom(opts prometheus.GaugeOpts, labelNames []string) metrics.Gauge {
	gv := prometheus.NewGaugeVec(opts, labelNames)
	prometheus.MustRegister(gv)
	return NewGauge(gv)
}

// NewGauge new a prometheus gauge and returns Gauge.
func NewGauge(gv *prometheus.GaugeVec) metrics.Gauge {
	return &gauge{
		gv: gv,
	}
}

func (g *gauge) With(lvs ...string) metrics.Gauge {
	return &gauge{
		gv:  g.gv,
		lvs: g.lvs.With(lvs...),
	}
}

func (g *gauge) Set(value float64) {
	g.gv.With(makeLabels(g.lvs...)).Set(value)
}

func (g *gauge) Add(delta float64) {
	g.gv.With(makeLabels(g.lvs...)).Add(delta)
}

func (g *gauge) Sub(delta float64) {
	g.gv.With(makeLabels(g.lvs...)).Sub(delta)
}
