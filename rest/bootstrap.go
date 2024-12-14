/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package rest

import (
	"github.com/siddontang/go-log/log"
	"github.com/jhblack-olya/trade-engine/conf"
)

// StartServer for rest server initialization
func StartServer() {
	gbeConfig := conf.GetConfig()

	httpServer := NewHttpServer(gbeConfig.RestServer.Addr)
	go httpServer.Start()

	log.Info("rest server ok")

	wsServer := NewWsServer(gbeConfig.WSserver.Addr)
	go wsServer.Start()
	log.Info("websocket server ok")

}
