package main

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"time"

	stdlog "log"

	"github.com/alecthomas/kingpin/v2"
	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
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
	toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9116")
	metricsPath  = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	includeExporterMetrics = kingpin.Flag(
		"web.include-exporter-metrics",
		"Include metrics about the exporter itself (promhttp_*, process_*, go_*) (default: disabled).",
	).Bool()
	maxRequests = kingpin.Flag(
		"web.max-requests",
		"Maximum number of parallel scrape requests. Use 0 to disable.",
	).Default("40").Int()
	maxProcs = kingpin.Flag(
		"runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)",
	).Envar("GOMAXPROCS").Default("1").Int()
	uri = kingpin.Flag(
		"uri",
		"clickhouse address. http://ip:http_port or tcp://ip:tcp_port",
	).Default("http://127.0.0.1:8123").String()
	user = kingpin.Flag(
		"user",
		"user for clickhouse",
	).Envar("CLICKHOUSE_USER").Short('u').String()
	password = kingpin.Flag(
		"password",
		"password for clickhouse",
	).Envar("CLICKHOUSE_PASSWORD").Short('p').String()
	maxExecutionTime = kingpin.Flag(
		"max-execution-time",
		"clickhouse client option MaxExecutionTime in seconds",
	).Default("120").Int()
	maxIdleConns = kingpin.Flag(
		"max-idle-conns",
		"clickhouse client option MaxIdleConns",
	).Default("5").Int()
	maxOpenConns = kingpin.Flag(
		"max-open-conns",
		"clickhouse client option maxOpenConns",
	).Default("5").Int()
	dialTimeout = kingpin.Flag(
		"dial-timeout",
		"clickhouse client option Dialtimeout in seconds",
	).Default("30").Int()
	connMaxLifetime = kingpin.Flag(
		"conn-max-lifetime",
		"clickhouse client option ConnMaxLifetime in minutes",
	).Default("60").Int()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("clickhouse_exporter"))
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)
	level.Info(logger).Log("msg", "Starting clickhouse_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	runtime.GOMAXPROCS(*maxProcs)
	level.Debug(logger).Log("msg", "Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	var opt = db.Option{
		MaxExecutionTime: *maxExecutionTime,
		MaxIdleConns:     *maxIdleConns,
		MaxOpenConns:     *maxOpenConns,
		ConnMaxLifetime:  time.Duration(*connMaxLifetime) * time.Minute,
		DialTimeout:      time.Duration(*dialTimeout) * time.Second,
	}
	u, err := url.Parse(*uri)
	if err != nil {
		stdlog.Fatal(err)
	}
	switch u.Scheme {
	case "tcp":
		db.InitTCPClient(u.Host, *user, *password, opt)
	case "http":
		db.InitHTTPClient(u.Host, *user, *password, opt)
	default:
		stdlog.Fatal("unexpected clickhouse uri: " + *uri)
	}
	level.Info(logger).Log("Scraping", *uri)

	if err := db.Ping(); err != nil {
		level.Error(logger).Log("msg", "clickhouse ping error", "err", err)
	} else {
		level.Info(logger).Log("msg", "clickhouse ping success")
	}

	e := exporter.NewExporter(logger)
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
			stdlog.Fatal(err)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		stdlog.Fatal(err)
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
