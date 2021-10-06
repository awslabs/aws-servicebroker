package broker

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// MetricsCollector is a prometheus metrics collector that is capable
// of providing better (more fine grained) action counts that the
// OSBMetricsCollector provided by the osb-broker-lib library - mainly
// it exists so that we don't reuse the same metric name, and we don't
// conflict with the metric gathering in that library.
type MetricsCollector struct {
	Actions *prom.CounterVec
}


// New initialises the MetricsCollector with a counter vec.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		Actions: prom.NewCounterVec(prom.CounterOpts{
			Name: "aws_sb_actions_total",
			Help: "Total amount of OpenServiceBroker actions requested.",
		}, []string{"action", "service", "plan"}),
	}
}

// Describe returns all descriptions of the collector.
func (c *MetricsCollector) Describe(ch chan<- *prom.Desc) {
	c.Actions.Describe(ch)
}

// Collect returns the current state of all metrics of the collector.
func (c *MetricsCollector) Collect(ch chan<- prom.Metric) {
	c.Actions.Collect(ch)
}



