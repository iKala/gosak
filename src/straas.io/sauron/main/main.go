package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"

	"straas.io/sauron"
	"straas.io/sauron/core"
	"straas.io/sauron/plugin/metric"
)

var (
	dryRun       = flag.Bool("dryRun", false, "dryrun mode")
	tickInterval = flag.Duration("jobTicker", time.Minute, "job runner ticker")
	esHosts      = flag.String("esHosts", "", "elasticsearch url list separarted in comma")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TODO: log util
// TODO: config loader
func main() {
	flag.Parse()

	// create es client
	esClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(strings.Split(*esHosts, ",")...),
		elastic.SetMaxRetries(10))
	if err != nil {
		log.Fatalf("fail to creat elasticsearch client, err:%v", err)
	}

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
			Env:      "*",
			Script: `
			a = mquery("revealer-syncer", "syncjob.proc_time", "value", "sum", "15m", "1m");
			console.log(a);
			`,
		},
	}
	// list all plugin
	plugins := []sauron.Plugin{
		metric.NewQuery(esClient),
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

	// TODO: better handling (slack)
	go func() {
		for e := range runner.Events() {
			fmt.Println("runner event %v", e)
		}
	}()

	// TODO: Leader election for cluster
	// TODO: add plugin help msg to "go help" message
	// TODO: listen http for health check
	time.Sleep(time.Minute)
}
