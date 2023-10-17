package collector

import (
	"fmt"
	"strconv"

	"github.com/ClickHouse/clickhouse_exporter/util"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newQueryDurationCollector(), true)
}

func newQueryDurationCollector() Collector {
	return &queryDurationCollector{}
}

type queryDurationCollector struct {
}

func (c *queryDurationCollector) Name() string {
	return "query_duration"
}

func (c *queryDurationCollector) Query() string {
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

func (c *queryDurationCollector) Collect(ch chan<- prometheus.Metric, data []byte) error {
	queryDurationMetrics, err := parseQueryResponse(data)
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for i, m := range queryDurationMetrics {
		newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "query_duration_ms",
			Help:      "The number of milliseconds spent on query.",
		}, []string{"sql", "top"}).WithLabelValues(m.key, strconv.Itoa(i+1))
		newMetric.Set(m.value)
		newMetric.Collect(ch)
	}

	return nil
}
