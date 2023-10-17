package exporter

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "clickhouse" // For Prometheus metrics.
)

// Exporter collects clickhouse stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	client *http.Client

	scrapeFailures prometheus.Counter

	user     string
	password string

	collectors          []collector.Collector
	collectorMetricURIs map[collector.Collector]string
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri url.URL, insecure bool, user, password string) *Exporter {
	var collectors = []collector.Collector{}
	var collectorMetricURIs = map[collector.Collector]string{}
	for v, enable := range collector.Collectors {
		if *enable {
			collectors = append(collectors, v)

			q := uri.Query()
			q.Set("query", v.Query())
			uri.RawQuery = q.Encode()
			collectorMetricURIs[v] = uri.String()
		}
	}

	return &Exporter{
		collectors:          collectors,
		collectorMetricURIs: collectorMetricURIs,
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping clickhouse.",
		}),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
			},
			Timeout: 30 * time.Second,
		},
		user:     user,
		password: password,
	}
}

// Describe describes all the metrics ever exported by the clickhouse exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(e, ch)
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	for _, c := range e.collectors {
		data, err := e.handleResponse(c)
		if err != nil {
			return err
		}
		err = c.Collect(ch, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Exporter) handleResponse(c collector.Collector) ([]byte, error) {
	req, err := http.NewRequest("GET", e.collectorMetricURIs[c], nil)
	if err != nil {
		return nil, err
	}

	// https://clickhouse.com/docs/zh/interfaces/formats#json
	// req.Header.Set("X-ClickHouse-Format", "TabSeparated")
	// if c.Name() == "query_duration" || c.Name() == "query_memory" {
	// 	req.Header.Set("X-ClickHouse-Format", "JSON")
	// }

	if e.user != "" && e.password != "" {
		req.Header.Set("X-ClickHouse-User", e.user)
		req.Header.Set("X-ClickHouse-Key", e.password)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error scraping clickhouse: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if err != nil {
			data = []byte(err.Error())
		}
		return nil, fmt.Errorf("status %s (%d): %s", resp.Status, resp.StatusCode, data)
	}

	return data, nil
}

// Collect fetches the stats from configured clickhouse location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	upValue := 1

	if err := e.collect(ch); err != nil {
		log.Printf("Error scraping clickhouse: %s", err)
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
