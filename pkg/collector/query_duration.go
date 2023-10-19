package collector

import (
	"fmt"
	"strconv"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
	"github.com/fuxingZhang/clickhouse_exporter/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newQueryDurationCollector(), true)
}

func newQueryDurationCollector() Collector {
	return &queryDurationCollector{}
}

type queryDurationCollector struct{}

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
	metrics, err := db.GetKeyValueData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for i, v := range metrics {
		newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "query_duration_ms",
			Help:      "The number of milliseconds spent on query.",
		}, []string{"sql", "top"}).WithLabelValues(v.Key, strconv.Itoa(i+1))
		newMetric.Set(v.Val)
		newMetric.Collect(ch)
	}

	return nil
}
