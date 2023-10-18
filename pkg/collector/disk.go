package collector

import (
	"fmt"

	"github.com/fuxingZhang/clickhouse_exporter/pkg/db"
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

func (c *diskCollector) SQL() string {
	return `select name, sum(free_space) as free_space_in_bytes, sum(total_space) as total_space_in_bytes from system.disks group by name`
}

func (c *diskCollector) Collect(ch chan<- prometheus.Metric) error {
	metrics, err := db.GetDiskData(c.SQL())
	if err != nil {
		return fmt.Errorf("error scraping clickhouse collector %v: %v", c.Name(), err)
	}

	for _, v := range metrics {
		newFreeSpaceMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "free_space_in_bytes",
			Help:      "Disks free_space_in_bytes capacity",
		}, []string{"disk"}).WithLabelValues(v.Disk)
		newFreeSpaceMetric.Set(v.FreeSpace)
		newFreeSpaceMetric.Collect(ch)

		newTotalSpaceMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "total_space_in_bytes",
			Help:      "Disks total_space_in_bytes capacity",
		}, []string{"disk"}).WithLabelValues(v.Disk)
		newTotalSpaceMetric.Set(v.TotalSpace)
		newTotalSpaceMetric.Collect(ch)
	}

	return nil
}
