package collector

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(newDiskCollector(), true)
}

func newDiskCollector() Collector {
	return &diskCollector{}
}

type diskCollector struct {
}

func (c *diskCollector) Name() string {
	return "disk"
}

func (c *diskCollector) Query() string {
	return `select name, sum(free_space) as free_space_in_bytes, sum(total_space) as total_space_in_bytes from system.disks group by name`
}

func (c *diskCollector) Collect(ch chan<- prometheus.Metric, data []byte) error {
	disksMetrics, err := c.parseDiskResponse(data)
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, dm := range disksMetrics {
		newFreeSpaceMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "free_space_in_bytes",
			Help:      "Disks free_space_in_bytes capacity",
		}, []string{"disk"}).WithLabelValues(dm.disk)
		newFreeSpaceMetric.Set(dm.freeSpace)
		newFreeSpaceMetric.Collect(ch)

		newTotalSpaceMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "total_space_in_bytes",
			Help:      "Disks total_space_in_bytes capacity",
		}, []string{"disk"}).WithLabelValues(dm.disk)
		newTotalSpaceMetric.Set(dm.totalSpace)
		newTotalSpaceMetric.Collect(ch)
	}

	return nil
}

type diskResult struct {
	disk       string
	freeSpace  float64
	totalSpace float64
}

func (c *diskCollector) parseDiskResponse(data []byte) ([]diskResult, error) {
	// Parsing results
	lines := strings.Split(string(data), "\n")
	var results []diskResult = make([]diskResult, 0)

	for i, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		if len(parts) != 3 {
			return nil, fmt.Errorf("parseDiskResponse: unexpected %d line: %s", i, line)
		}
		disk := strings.TrimSpace(parts[0])

		freeSpace, err := parseNumber(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}

		totalSpace, err := parseNumber(strings.TrimSpace(parts[2]))
		if err != nil {
			return nil, err
		}

		results = append(results, diskResult{disk, freeSpace, totalSpace})

	}
	return results, nil
}
