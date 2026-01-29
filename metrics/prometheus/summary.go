package prometheus

import (
	"github.com/iWuxc/go-wit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var _ metrics.Observer = (*summary)(nil)

type summary struct {
	sv  *prometheus.SummaryVec
	lvs LabelValues
}

// NewSummaryFrom constructs and registers a Prometheus SummaryVec,
// and returns a usable Summary object.
func NewSummaryFrom(opts prometheus.SummaryOpts, labelNames []string) metrics.Observer {
	sv := prometheus.NewSummaryVec(opts, labelNames)
	prometheus.MustRegister(sv)
	return NewSummary(sv)
}

// NewSummary new a prometheus summary and returns Histogram.
func NewSummary(sv *prometheus.SummaryVec) metrics.Observer {
	return &summary{
		sv: sv,
	}
}

func (s *summary) With(lvs ...string) metrics.Observer {
	return &summary{
		sv:  s.sv,
		lvs: s.lvs.With(lvs...),
	}
}

func (s *summary) Observe(value float64) {
	s.sv.With(makeLabels(s.lvs...)).Observe(value)
}
