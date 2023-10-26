package collector

import (
	"fmt"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newPartsCollector(), true)
}

func newPartsCollector() Collector {
	const subsystem = "table_parts"

	return &partsCollector{
		bytesMetricVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "bytes",
			Help:      "Table size in bytes",
		}, []string{"database", "table"}),
		countMetricVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "count",
			Help:      "Number of parts of the table",
		}, []string{"database", "table"}),
		rowsMetricVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "rows",
			Help:      "Number of rows in the table",
		}, []string{"database", "table"}),
	}
}

type partsCollector struct {
	bytesMetricVec *prometheus.GaugeVec
	countMetricVec *prometheus.GaugeVec
	rowsMetricVec  *prometheus.GaugeVec
}

func (c *partsCollector) Name() string {
	return "parts"
}

func (c *partsCollector) SQL() string {
	return `select database, table, sum(bytes) as bytes, count() as parts, sum(rows) as rows from system.parts where active = 1 group by database, table`
}

func (c *partsCollector) Collect(ch chan<- prometheus.Metric) error {
	metrics, err := db.GetPartsData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, v := range metrics {
		labels := prometheus.Labels{
			"database": v.Database,
			"table":    v.Table,
		}

		bytesMetric := c.bytesMetricVec.With(labels)
		bytesMetric.Set(float64(v.Bytes))
		bytesMetric.Collect(ch)

		countMetric := c.countMetricVec.With(labels)
		countMetric.Set(float64(v.Parts))
		countMetric.Collect(ch)

		rowsMetric := c.rowsMetricVec.With(labels)
		rowsMetric.Set(float64(v.Rows))
		rowsMetric.Collect(ch)
	}

	return nil
}
