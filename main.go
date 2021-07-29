package main

import (
	"github.com/prometheus/common/log"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/rest"

	"net/http"
	_ "net/http/pprof"
)

func main() {

	go func() {
		log.Info(http.ListenAndServe("localhost:6000", nil))
	}()

	matching.StartEngine()
	rest.StartServer()

	select {}
}
