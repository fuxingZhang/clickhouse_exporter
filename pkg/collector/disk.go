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
	type DiskData struct {
		Disk       string  `gorm:"column:name"`
		FreeSpace  float64 `gorm:"column:free_space_in_bytes"`
		TotalSpace float64 `gorm:"column:total_space_in_bytes"`
	}

	var metrics []DiskData

	err := db.DB.Raw(c.SQL()).Scan(&metrics).Error
	// err := db.DB.Raw(c.SQL()).Find(&metrics).Error
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
