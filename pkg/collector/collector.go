package collector

import (
	"fmt"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "clickhouse" // For Prometheus metrics.
)

const (
	defaultEnabled  = true
	defaultDisabled = false
)

// Collectors Collectors
var Collectors = map[Collector]*bool{}

// Collector is the interface a collector has to implement.
type Collector interface {
	Name() string
	Query() string
	// Get new metrics and expose them via prometheus registry.
	Collect(ch chan<- prometheus.Metric, data []byte) error
}

func registerCollector(collector Collector, isDefaultEnabled bool) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector.Name())
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector.Name(), helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Bool()
	Collectors[collector] = flag
}
