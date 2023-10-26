package collector

import (
	"fmt"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newEventsCollector(), true)
}

func newEventsCollector() Collector {
	return &eventsCollector{}
}

type eventsCollector struct {
}

func (c *eventsCollector) Name() string {
	return "events"
}

func (c *eventsCollector) SQL() string {
	return `select event, value from system.events`
}

func (c *eventsCollector) Collect(ch chan<- prometheus.Metric) error {
	events, err := db.GetKeyValueData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, v := range events {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				namespace+"_"+metricName(v.Key)+"_total",
				"Number of "+v.Key+" total processed",
				[]string{}, nil),
			prometheus.CounterValue, v.Val)
	}

	return nil
}
