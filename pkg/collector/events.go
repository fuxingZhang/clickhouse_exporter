package collector

import (
	"fmt"

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

func (c *eventsCollector) Query() string {
	return `select event, value from system.events`
}

func (c *eventsCollector) Collect(ch chan<- prometheus.Metric, data []byte) error {
	events, err := parseKeyValueResponse(data)
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, ev := range events {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				namespace+"_"+metricName(ev.key)+"_total",
				"Number of "+ev.key+" total processed", []string{}, nil),
			prometheus.CounterValue, float64(ev.value))
	}

	return nil
}
