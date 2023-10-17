package main

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"

	stdlog "log"

	"github.com/alecthomas/kingpin/v2"
	"github.com/fuxingZhang/clickhouse_exporter/pkg/exporter"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	includeExporterMetrics = kingpin.Flag(
		"web.include-exporter-metrics",
		"Include metrics about the exporter itself (promhttp_*, process_*, go_*).",
	).Bool()
	maxRequests = kingpin.Flag(
		"web.max-requests",
		"Maximum number of parallel scrape requests. Use 0 to disable.",
	).Default("40").Int()
	maxProcs = kingpin.Flag(
		"runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)",
	).Envar("GOMAXPROCS").Default("1").Int()
	toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9116")

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

	logger := promlog.New(promlogConfig)
	level.Info(logger).Log("msg", "Starting clickhouse_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	uri, err := url.Parse(*clickhouseScrapeURI)
	if err != nil {
		panic(err)
	}
	level.Info(logger).Log("Scraping", *clickhouseScrapeURI)

	// registerer := prometheus.DefaultRegisterer
	// gatherer := prometheus.DefaultGatherer
	// if *clickhouseOnly {
	// 	reg := prometheus.NewRegistry()
	// 	registerer = reg
	// 	gatherer = reg
	// }

	e := exporter.NewExporter(*uri, *insecure, *user, *password, logger)
	// registerer.MustRegister(e)

	// http.Handle(*metricsPath, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))

	runtime.GOMAXPROCS(*maxProcs)
	level.Debug(logger).Log("msg", "Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	http.Handle(*metricsPath, handler(*includeExporterMetrics, *maxRequests, e, logger))

	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        "Clickhouse Exporter",
			Description: "Prometheus Clickhouse Exporter",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}

func handler(includeExporterMetrics bool, maxRequests int, c prometheus.Collector, logger log.Logger) http.Handler {
	registry := prometheus.NewRegistry()

	registry.MustRegister(version.NewCollector("clickhouse_exporter"))

	registry.MustRegister(collectors.NewBuildInfoCollector())

	if includeExporterMetrics {
		registry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(
				collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
			),
		)
	}

	registry.MustRegister(c)

	handler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: maxRequests,
			// promhttp_metric_handler_errors_total
			Registry: registry,
		},
	)

	if includeExporterMetrics {
		handler = promhttp.InstrumentMetricHandler(
			registry, handler,
		)
	}

	return handler
}
