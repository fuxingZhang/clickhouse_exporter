package collector

import (
	"fmt"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
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

func (c *asyncMetricsCollector) SQL() string {
	return `select replaceRegexpAll(toString(metric), '-', '_') AS metric, value from system.asynchronous_metrics`
}

func (c *asyncMetricsCollector) Collect(ch chan<- prometheus.Metric) error {
	metrics, err := db.GetKeyValueData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, v := range metrics {
		newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName(v.Key),
			Help:      "Number of " + v.Key + " async processed",
		}, []string{}).WithLabelValues()
		newMetric.Set(v.Val)
		newMetric.Collect(ch)
	}

	return nil
}
