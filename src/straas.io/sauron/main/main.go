package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"

	"straas.io/base/ctrl"
	"straas.io/base/logger"
	"straas.io/base/timeutil"
	"straas.io/sauron"
	"straas.io/sauron/core"
	// externals
	"straas.io/external/pagerduty"
	"straas.io/external/slack"
	// plugins
	"straas.io/sauron/plugin/alert"
	"straas.io/sauron/plugin/metric"
	"straas.io/sauron/plugin/notification"
	// notification sinkers
	ntyPagerDuty "straas.io/sauron/plugin/notification/pagerduty"
	ntySlack "straas.io/sauron/plugin/notification/slack"
)

var (
	configRoot   = flag.String("configRoot", "config/", "config root folder")
	dryRun       = flag.Bool("dryRun", true, "dryrun mode")
	envStr       = flag.String("envs", "", "environments separated by comma")
	tickInterval = flag.Duration("jobTicker", time.Minute, "job runner ticker")
	esHosts      = flag.String("esHosts", "", "elasticsearch url list separarted in comma")
	jobPattern   = flag.String("jobPattern", "", "only run jobs matches the pattern, only for dryrun mode")
	logLevel     = flag.String("logLevel", "info", "log level")
	port         = flag.Int("port", 8000, "port for health check")
	log          = logger.Get()
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	flag.Parse()

	// setup default time for default http client
	http.DefaultClient.Timeout = 30 * time.Second

	// create logger
	if err := logger.SetLevel(*logLevel); err != nil {
		log.Fatalf("illegal log level %s", *logLevel)
	}
	// create clock
	clock := timeutil.NewRealClock()
	// create output
	output := core.NewOutput(*dryRun)

	// parse environemnts
	envs := strings.Split(*envStr, ",")
	log.Info("[main] environemnt", envs)

	cfgMgr, err := core.NewFileConfig(*configRoot, *dryRun)
	if err != nil {
		log.Fatalf("[main] fail to load create config manager, err:%v", err)
	}

	// create es client
	esClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(strings.Split(*esHosts, ",")...),
		elastic.SetMaxRetries(10))
	if err != nil {
		log.Fatalf("[main] fail to creat elasticsearch client, err:%v", err)
	}

	// create store
	statusStore, err := core.NewStore()
	if err != nil {
		log.Fatalf("[main] fail to init store, err:%v", err)
	}

	// prepare ticker
	var ticker <-chan time.Time
	// dry run only need to tick once immediately
	if !*dryRun {
		ticker = time.NewTicker(*tickInterval).C
	}

	// prepare engine factory
	engFactory := func() sauron.Engine {
		return core.NewEngine(statusStore, output)
	}

	// load jobs
	jobs, err := cfgMgr.LoadJobs(envs...)
	if err != nil {
		log.Fatalf("[main] fail to load jobs, err:%v", err)
	}

	// register sinkers
	// TODO: es insert sinker
	log.Info("[main] register notification sinkers")
	notification.RegisterSinker("slack", func() notification.Sinker {
		return ntySlack.NewSinker(slack.New())
	})
	notification.RegisterSinker("pagerduty", func() notification.Sinker {
		return ntyPagerDuty.NewSinker(pagerduty.New(), clock)
	})

	// create notifiation
	ntyPlugin, err := notification.NewNotification(cfgMgr)
	if err != nil {
		log.Fatalf("[main] fail to init notification plugin, err:%v", err)
	}

	// list all plugin
	log.Info("[main] register plugins")
	plugins := []sauron.Plugin{
		alert.NewLastFor(clock),
		alert.NewAlert(clock),
		metric.NewQuery(esClient, clock),
		ntyPlugin,
	}
	for _, p := range plugins {
		log.Infof("[main] register plugin %s", p.Name())
	}

	// runnerID
	runnerID := fmt.Sprintf("RN%d", rand.Int63())
	log.Infof("[main] Sauron runner id %s", runnerID)

	// create runner
	runner := core.NewJobRunner(
		runnerID,
		engFactory,
		ticker,
		statusStore,
		output,
		jobs,
		plugins,
		clock,
	)

	log.Info("[main] start to run jobs")
	runner.Start()

	// TODO: better handling (slack)
	go func() {
		for e := range runner.Events() {
			log.Errorf("[main] runner event %v", e)
		}
	}()

	// TODO: Leader election for cluster (if necessary)
	// TODO: add plugin help msg to "go help" message
	// TODO: handle graceful shutdown
	if *dryRun {
		for _, j := range jobs {
			if *jobPattern != "" && !strings.Contains(j.JobID, *jobPattern) {
				continue
			}
			if err := runner.RunJob(j); err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	log.Fatal(ctrl.RunController(*port))
}
