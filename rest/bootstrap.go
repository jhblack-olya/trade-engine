package rest

import (
	"github.com/siddontang/go-log/log"
	"gitlab.com/gae4/trade-engine/conf"
)

//StartServer for rest server initialization
func StartServer() {
	gbeConfig := conf.GetConfig()

	httpServer := NewHttpServer(gbeConfig.RestServer.Addr)
	go httpServer.Start()

	log.Info("rest server ok")
}
