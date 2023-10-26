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
	return &partsCollector{}
}

type partsCollector struct {
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
		bytesMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "table_parts_bytes",
			Help:      "Table size in bytes",
		}, []string{"database", "table"}).WithLabelValues(v.Database, v.Table)
		bytesMetric.Set(float64(v.Bytes))
		bytesMetric.Collect(ch)

		countMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "table_parts_count",
			Help:      "Number of parts of the table",
		}, []string{"database", "table"}).WithLabelValues(v.Database, v.Table)
		countMetric.Set(float64(v.Parts))
		countMetric.Collect(ch)

		rowsMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "table_parts_rows",
			Help:      "Number of rows in the table",
		}, []string{"database", "table"}).WithLabelValues(v.Database, v.Table)
		rowsMetric.Set(float64(v.Rows))
		rowsMetric.Collect(ch)
	}

	return nil
}
