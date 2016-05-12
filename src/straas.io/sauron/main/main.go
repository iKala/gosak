package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"straas.io/sauron"
	"straas.io/sauron/core"
	"straas.io/sauron/plugin/metric"
)

var (
	dryRun       = flag.Bool("dryRun", false, "dryrun mode")
	tickInterval = flag.Duration("jobTicker", time.Minute, "job runner ticker")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TODO: log util
// TODO: config loader
func main() {
	flag.Parse()

	// create store
	statusStore, err := core.NewStore()
	if err != nil {
		log.Fatalf("fail to init store, err:%v", err)
	}
	// prepare ticker
	var ticker <-chan time.Time
	// dry run only need to tick once immediately
	if *dryRun {
		oneTimeTicker := make(chan time.Time, 1)
		ticker = oneTimeTicker
		oneTimeTicker <- time.Now()
	} else {
		ticker = time.NewTicker(*tickInterval).C
	}

	// prepare engine factory
	engFactory := core.NewEngine
	// read jobs
	jobs := []sauron.JobMeta{
		sauron.JobMeta{
			JobID:    "aaaaa",
			Interval: time.Minute,
			Script:   `console.log(mquery(4, 5))`,
		},
	}
	// list all plugin
	plugins := []sauron.Plugin{
		metric.NewQuery(),
	}

	// runnerID
	runnerID := fmt.Sprintf("RN%d", rand.Int63())
	log.Printf("Sauron runner id %s", runnerID)

	// create runner
	runner := core.NewJobRunner(
		runnerID,
		engFactory,
		ticker,
		statusStore,
		jobs,
		plugins)

	runner.Start()
	log.Println("start to run jobs")

	// TODO: better handling
	go func() {
		for e := range runner.Events() {
			fmt.Println("runner event %v", e)
		}
	}()

	// TODO: listen http for health check
	time.Sleep(time.Minute)
}
