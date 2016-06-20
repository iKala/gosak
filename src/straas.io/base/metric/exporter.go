package metric

import (
	"fmt"
	"os"
	"time"

	"github.com/csigo/metric"

	"straas.io/external"
)

const (
	getInterval   = 2 * time.Minute
	checkInterval = 15 * time.Second
)

var (
	hostName, _ = os.Hostname()
	percentiles = []float64{0.5, 0.95, 0.99}
)

// StartExport start export metrics
func StartExport(fluent external.Fluent, tag string, done chan bool) {
	go runExport(fluent, tag, done)
}

// exporter define export type
type exporter func(pkg, name string, count, sum, avg float64)

// for test
var getSnapshot = metric.GetSnapshot

func runExport(fluent external.Fluent, tag string, done chan bool) {
	lastUpdate := time.Unix(0, 0)
	exp := createExporter(fluent, tag)
	for {
		lastUpdate = exportOnce(lastUpdate, exp)
		time.Sleep(checkInterval)
	}
}

func createExporter(fluent external.Fluent, tag string) exporter {
	return func(pkg, name string, count, sum, avg float64) {
		fluent.Post(tag, map[string]interface{}{
			"host":   hostName,
			"module": pkg,
			"name":   name,
			"value":  sum,
			"avg":    avg,
			"count":  count,
		})
	}
}

func exportOnce(lastUpdated time.Time, exp exporter) time.Time {
	nextUpdate := lastUpdated

	for _, s := range getSnapshot("*", "*") {
		bks := s.SliceIn(getInterval)
		if len(bks) == 0 {
			continue
		}
		bk := bks[len(bks)-1]
		if !bk.End.After(lastUpdated) {
			continue
		}
		nextUpdate = bk.End

		// export counter
		exp(s.Pkg(), s.Name(), bk.Count, bk.Sum, bk.Avg)

		// export histogram percentiles
		if !s.HasHistogram() {
			continue
		}
		pvs, count := s.Percentiles(percentiles)
		if len(pvs) == 0 {
			continue
		}
		for i, pv := range pvs {
			pInt := int(percentiles[i] * 100)
			exp(s.Pkg(), fmt.Sprintf("%s.p%d", s.Name(), pInt), float64(count), pv*float64(count), pv)
		}
	}
	return nextUpdate
}
