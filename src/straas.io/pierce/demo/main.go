package main

import (
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/coreos/etcd/client"

	"straas.io/base/logger"
	"straas.io/pierce/core"
	"straas.io/pierce/socket"
)

var (
	log = logger.Get()
)

// this is only the demo code for socket server
// should be merged into main later
func main() {
	flag.Parse()

	DefaultTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}

	cfg := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: DefaultTransport,
	}

	c, err := client.New(cfg)
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	kAPI := client.NewKeysAPI(c)
	coreMgr := core.NewCore(kAPI)
	coreMgr.Start()

	skServer := socket.NewSockerServer(coreMgr)
	handler, err := skServer.Create()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/socket.io/", handler)
	log.Info("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
