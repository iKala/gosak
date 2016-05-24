package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	elastic "gopkg.in/olivere/elastic.v3"

	"straas.io/base/ctrl"
	"straas.io/base/logger"
	"straas.io/base/timeutil"
	"straas.io/sauron"
	"straas.io/sauron/core"
	"straas.io/sauron/program"
	// externals
	"straas.io/external"
	"straas.io/external/elasticsearch"
	"straas.io/external/pagerduty"
	"straas.io/external/slack"
	estackdrive "straas.io/external/stackdriver"
	// plugins
	"straas.io/sauron/plugin/alert"
	"straas.io/sauron/plugin/metric"
	"straas.io/sauron/plugin/notification"
	"straas.io/sauron/plugin/stackdriver"
	// notification sinkers
	ntyPagerDuty "straas.io/sauron/plugin/notification/pagerduty"
	ntySlack "straas.io/sauron/plugin/notification/slack"
)

const (
	cacheSize = 3000
)

var (
	configRoot   = flag.String("configRoot", "config/sauron", "config root folder")
	dryRun       = flag.Bool("dryRun", true, "dryrun mode")
	envStr       = flag.String("envs", "", "environments separated by comma")
	tickInterval = flag.Duration("jobTicker", time.Minute, "job runner ticker")
	jobPattern   = flag.String("jobPattern", "", "only run jobs matches the pattern, only for dryrun mode")
	logLevel     = flag.String("logLevel", "info", "log level")
	port         = flag.Int("port", 8000, "port for health check")
	log          = logger.Get()
	// flags for plugins
	esHosts = flag.String("esHosts", "", "elasticsearch url list separarted in comma")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func initMetricPlugin(p sauron.Program, esHosts string) error {
	// create es client
	esClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(strings.Split(esHosts, ",")...),
		elastic.SetMaxRetries(10))
	if err != nil {
		return err
	}
	clock := timeutil.NewRealClock()
	p.AddPlugin(metric.NewQuery(elasticsearch.New(esClient), clock))
	return nil
}

func initStackdrivers(p sauron.Program, envs []string, mainCfg Config) error {
	proj2sd := map[string]external.Stackdriver{}
	for _, env := range envs {
		project, ok := mainCfg.Env2GCP[env]
		if !ok {
			return fmt.Errorf("unable to find GCP project for env %s", env)
		}
		auth, ok := mainCfg.Credential.GCP[project]
		if !ok {
			return fmt.Errorf("unable to find credential for GCP project %s", project)
		}

		ctx := context.Background()
		jwtCfg := &jwt.Config{
			Email:      auth.Email,
			PrivateKey: []byte(auth.PrivateKey),
			TokenURL:   google.JWTTokenURL,
			Scopes:     auth.Scopes,
		}
		client := jwtCfg.Client(ctx)
		sd, err := estackdrive.New(client)
		if err != nil {
			return err
		}
		proj2sd[project] = sd
	}
	clock := timeutil.NewRealClock()
	p.AddPlugin(stackdriver.NewQuery(proj2sd, mainCfg.Env2GCP, clock))
	return nil
}

func initAlertPlugin(p sauron.Program) error {
	clock := timeutil.NewRealClock()
	p.AddPlugin(alert.NewLastFor(clock))
	p.AddPlugin(alert.NewAlert(clock))
	return nil
}

func initNtyPlugin(p sauron.Program, cfgMgr sauron.Config, mainCfg Config) error {
	credential := mainCfg.Credential
	clock := timeutil.NewRealClock()

	// register sinkers
	// TODO: es insert sinker
	log.Info("[main] register notification sinkers")
	notification.RegisterSinker("slack", func() notification.Sinker {
		return ntySlack.NewSinker(slack.New(credential.SlackToken))
	})
	notification.RegisterSinker("pagerduty", func() notification.Sinker {
		return ntyPagerDuty.NewSinker(pagerduty.New(credential.PagerDutyToken), clock)
	})

	// create notifiation
	ntyPlugin, err := notification.NewNotification(cfgMgr)
	if err != nil {
		return fmt.Errorf("[main] fail to init notification plugin, err:%v", err)
	}
	p.AddPlugin(ntyPlugin)
	return nil
}

func main() {
	flag.Parse()

	// setup default time for default http client
	http.DefaultClient.Timeout = 30 * time.Second

	// create logger
	if err := logger.SetLevel(*logLevel); err != nil {
		log.Fatalf("[main]illegal log level %s", *logLevel)
	}

	// envs
	envs := strings.Split(*envStr, ",")
	// config manager
	cfgMgr, err := core.NewFileConfig(*configRoot, *dryRun)
	if err != nil {
		log.Fatalf("[main] fail to load create config manager, err:%v", err)
	}

	// main config
	mainCfg := Config{}
	if err := cfgMgr.LoadConfig("sauron", &mainCfg); err != nil {
		log.Fatalf("[main] fial to load main config, err:%v", err)
	}

	// create program
	p, err := program.New(envs, *dryRun, *jobPattern, cfgMgr, *tickInterval)
	if err != nil {
		log.Fatalf("[main] fail to create program, err:%v", err)
	}

	log.Info("[main] register plugins")
	if err := initMetricPlugin(p, *esHosts); err != nil {
		log.Fatalf("[main] fail to init metric plugin, err:%v", err)
	}
	if err := initStackdrivers(p, envs, mainCfg); err != nil {
		log.Fatalf("[main] fail to init stackdriver plugin, err:%v", err)
	}
	if err := initNtyPlugin(p, cfgMgr, mainCfg); err != nil {
		log.Fatalf("[main] fail to init notification plugin, err:%v", err)
	}
	if err := initAlertPlugin(p); err != nil {
		log.Fatalf("[main] fail to init alert plugin, err:%v", err)
	}

	p.AddEventHandler(func(je sauron.JobEvent) {
		log.Errorf("[main] runner event %v", je)
	})

	// start program
	if err := p.Start(); err != nil {
		log.Fatalf("[main] fail to start program, err:%v", err)
	}

	// TODO: Leader election for cluster (if necessary)
	// TODO: add plugin help msg to "go help" message
	// TODO: handle graceful shutdown
	// TODO: dashboard

	// prepare controller
	if !*dryRun {
		log.Fatal(ctrl.RunController(*port))
	}
}
