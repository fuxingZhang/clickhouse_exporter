package collector

import (
	"fmt"
	"strconv"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
	"github.com/fuxingZhang/clickhouse_exporter/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newQueryMemoryCollector(), false)
}

func newQueryMemoryCollector() Collector {
	return &queryMemoryCollector{
		queryMemoryVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "query",
			Name:      "memory_usage_bytes",
			Help:      "The number of memory bytes used by query",
		}, []string{"sql", "top"}),
	}
}

type queryMemoryCollector struct {
	queryMemoryVec *prometheus.GaugeVec
}

func (c *queryMemoryCollector) Name() string {
	return "query_memory"
}

func (c *queryMemoryCollector) SQL() string {
	return util.FormatSQL(`
	SELECT
		query,
		max(memory_usage) as memory_usage
	FROM
		system.query_log
	WHERE
		event_time >= now() - INTERVAL 5 MINUTE
		AND query NOT LIKE 'SELECT query, max(%) as % FROM system.query_log%'
	GROUP BY
		query
	ORDER BY
		memory_usage DESC
	LIMIT 5
	`)
}

func (c *queryMemoryCollector) Collect(ch chan<- prometheus.Metric) error {
	queryMemoryMetrics, err := db.GetKeyValueData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for i, v := range queryMemoryMetrics {
		// metric := c.queryMemoryVec.WithLabelValues(v.Key, strconv.Itoa(i+1))
		metric := c.queryMemoryVec.With(prometheus.Labels{
			"sql": v.Key,
			"top": strconv.Itoa(i + 1),
		})
		metric.Set(v.Val)
		metric.Collect(ch)
	}

	return nil
}
