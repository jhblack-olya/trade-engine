package main

import (
	"github.com/prometheus/common/log"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/pushing"

	"net/http"
	_ "net/http/pprof"

	"gitlab.com/gae4/trade-engine/rest"
)

func main() {

	go func() {
		log.Info(http.ListenAndServe("localhost:6000", nil))
	}()

	go models.NewBinLogStream().Start()

	matching.StartEngine()

	pushing.StartServer()

	rest.StartServer()
	select {}
}
