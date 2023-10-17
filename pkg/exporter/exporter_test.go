package exporter

import (
	"net/url"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
)

func TestScrape(t *testing.T) {
	clickhouseURL, err := url.Parse("http://127.0.0.1:8123/")
	if err != nil {
		t.Fatal(err)
	}
	logger := promlog.New(&promlog.Config{})
	exporter := NewExporter(*clickhouseURL, false, "", "", logger)

	t.Run("Describe", func(t *testing.T) {
		ch := make(chan *prometheus.Desc)
		go func() {
			exporter.Describe(ch)
			close(ch)
		}()

		for range ch {
		}
	})

	t.Run("Collect", func(t *testing.T) {
		ch := make(chan prometheus.Metric)
		var err error
		go func() {
			err = exporter.collect(ch)
			if err != nil {
				panic("failed")
			}
			close(ch)
		}()

		for range ch {
		}
	})
}
