package main

import (
	"net/http"
	"net/url"
	"os"

	"github.com/ClickHouse/clickhouse_exporter/exporter"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/log"
)

var (
	listeningAddress = kingpin.Flag(
		"telemetry.address",
		"Address on which to expose metrics.",
	).Default(":9116").String()
	metricsEndpoint = kingpin.Flag(
		"telemetry.endpoint",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	clickhouseScrapeURI = kingpin.Flag(
		"scrape_uri",
		"URI to clickhouse http endpoint",
	).Default("http://localhost:8123/").String()
	clickhouseOnly = kingpin.Flag(
		"clickhouse_only",
		"Expose only Clickhouse metrics, not metrics from the exporter itself",
	).Default("false").Bool()
	insecure = kingpin.Flag(
		"insecure",
		"Ignore server certificate if using https",
	).Default("true").Bool()
	user = kingpin.Flag(
		"user",
		"user for clickhouse",
	).Envar("CLICKHOUSE_USER").String()
	password = kingpin.Flag(
		"password",
		"password for clickhouse",
	).Envar("CLICKHOUSE_PASSWORD").String()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("clickhouse_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	uri, err := url.Parse(*clickhouseScrapeURI)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Scraping %s", *clickhouseScrapeURI)

	registerer := prometheus.DefaultRegisterer
	gatherer := prometheus.DefaultGatherer
	if *clickhouseOnly {
		reg := prometheus.NewRegistry()
		registerer = reg
		gatherer = reg
	}

	e := exporter.NewExporter(*uri, *insecure, *user, *password)
	registerer.MustRegister(e)

	http.Handle(*metricsEndpoint, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Clickhouse Exporter</title></head>
			<body>
			<h1>Clickhouse Exporter</h1>
			<p><a href="` + *metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
