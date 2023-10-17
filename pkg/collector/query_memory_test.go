package collector

import "testing"

func TestQueryMemoryQuery(t *testing.T) {
	c := newQueryMemoryCollector()
	t.Log(c.Query())
}
