package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newAsyncMetricsCollector(), true)
}

func newAsyncMetricsCollector() Collector {
	return &asyncMetricsCollector{}
}

type asyncMetricsCollector struct {
}

func (c *asyncMetricsCollector) Name() string {
	return "async_metrics"
}

func (c *asyncMetricsCollector) Query() string {
	return `select replaceRegexpAll(toString(metric), '-', '_') AS metric, value from system.asynchronous_metrics`
}

func (c *asyncMetricsCollector) Collect(ch chan<- prometheus.Metric, data []byte) error {
	asyncMetrics, err := parseKeyValueResponse(data)
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, am := range asyncMetrics {
		newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName(am.key),
			Help:      "Number of " + am.key + " async processed",
		}, []string{}).WithLabelValues()
		newMetric.Set(am.value)
		newMetric.Collect(ch)
	}

	return nil
}
