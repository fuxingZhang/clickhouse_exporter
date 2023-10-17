package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newMetricsCollector(), true)
}

func newMetricsCollector() Collector {
	return &metricsCollector{}
}

type metricsCollector struct {
}

func (c *metricsCollector) Name() string {
	return "metrics"
}

func (c *metricsCollector) Query() string {
	return `select metric, value from system.metrics`
}

func (c *metricsCollector) Collect(ch chan<- prometheus.Metric, data []byte) error {
	metrics, err := parseKeyValueResponse(data)
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, m := range metrics {
		newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName(m.key),
			Help:      "Number of " + m.key + " currently processed",
		}, []string{}).WithLabelValues()
		newMetric.Set(m.value)
		newMetric.Collect(ch)
	}
	return nil
}
