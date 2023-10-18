package exporter

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "clickhouse" // For Prometheus metrics.
)

// Exporter collects clickhouse stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	collectors     []collector.Collector
	scrapeFailures prometheus.Counter
	logger         log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(logger log.Logger) *Exporter {
	var collectors = []collector.Collector{}

	for v, enable := range collector.Collectors {
		if *enable {
			collectors = append(collectors, v)
		}
	}

	return &Exporter{
		collectors: collectors,
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping clickhouse.",
		}),
		logger: logger,
	}
}

// Describe describes all the metrics ever exported by the clickhouse exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(e, ch)
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	for _, c := range e.collectors {
		err := c.Collect(ch)
		if err != nil {
			return err
		}
	}

	return nil
}

// Collect fetches the stats from configured clickhouse location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	upValue := 1

	if err := e.collect(ch); err != nil {
		level.Error(e.logger).Log("msg", "Error scraping target", "err", err)
		e.scrapeFailures.Inc()
		e.scrapeFailures.Collect(ch)

		upValue = 0
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Was the last query of ClickHouse successful.",
			nil, nil,
		),
		prometheus.GaugeValue, float64(upValue),
	)
}
