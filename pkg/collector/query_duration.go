package collector

import (
	"fmt"
	"strconv"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
	"github.com/fuxingZhang/clickhouse_exporter/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newQueryDurationCollector(), false)
}

func newQueryDurationCollector() Collector {
	return &queryDurationCollector{
		queryDurationVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "query",
			Name:      "duration_ms",
			Help:      "The number of milliseconds spent on query.",
		}, []string{"sql", "top"}),
	}
}

type queryDurationCollector struct {
	queryDurationVec *prometheus.GaugeVec
}

func (c *queryDurationCollector) Name() string {
	return "query_duration"
}

func (c *queryDurationCollector) SQL() string {
	return util.FormatSQL(`
	SELECT
		query,
		max(query_duration_ms) as query_duration_ms
	FROM
		system.query_log
	WHERE 
		event_time >= now() - INTERVAL 5 MINUTE
		AND query NOT LIKE 'SELECT query, max(%) as % FROM system.query_log%'
	GROUP BY
		query
	ORDER BY
		query_duration_ms DESC
	LIMIT 5
	`)
}

func (c *queryDurationCollector) Collect(ch chan<- prometheus.Metric) error {
	queryDurationMetrics, err := db.GetKeyValueData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for i, v := range queryDurationMetrics {
		// metric := c.queryDurationVec.WithLabelValues(v.Key, strconv.Itoa(i+1))
		metric := c.queryDurationVec.With(prometheus.Labels{
			"sql": v.Key,
			"top": strconv.Itoa(i + 1),
		})

		metric.Set(v.Val)
		metric.Collect(ch)
	}

	return nil
}
