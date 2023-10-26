package collector

import (
	"fmt"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
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

func (c *metricsCollector) SQL() string {
	return `select metric, value from system.metrics`
}

func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) error {
	metrics, err := db.GetKeyValueData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, v := range metrics {
		metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName(v.Key),
			Help:      "Number of " + v.Key + " currently processed",
		}, []string{}).WithLabelValues()
		metric.Set(v.Val)
		metric.Collect(ch)
	}
	return nil
}
