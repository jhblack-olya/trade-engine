/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package rest

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/log"
)

type HttpServer struct {
	addr string
}

type WsServer struct {
	addr string
}

func NewHttpServer(addr string) *HttpServer {
	return &HttpServer{
		addr: addr,
	}
}

func NewWsServer(addr string) *WsServer {
	return &WsServer{
		addr: addr,
	}
}

func (server *HttpServer) Start() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.Default()
	r.Use(setCROSOptions)

	private := r.Group("/", checkAPIkey())
	{
		//private.POST("/api/orders", PlaceOrderAPI)
		private.POST("/api/backendOrder", BackendOrder) //for testing purpose
		private.POST("/api/account/create", CreateAccount)
		private.PATCH("/api/account/update", UpdateAccount)
		private.GET("/api/estimate", EstimateAmount)
		private.GET("/api/websocket/close", CloseWebsocket)
	}
	r.GET("/health", healthCheck())
	r.GET("/orderbook", GetLiveOrderBook)

	err := r.Run(server.addr)
	if err != nil {
		panic(err)
	}
}
func (server *WsServer) Start() {
	// router = gin.Default()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.Default()
	r.Use(setCROSOptions)

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	sockets := r.Group("/", checkAPIkey())
	{
		sockets.GET("/api/ws/orderbook", GetLiveOrderBook)
	}
	log.Info("Websocket starting at :: " + server.addr)

	err := r.Run(server.addr)
	if err != nil {
		panic(err)
	}

}

func setCROSOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Content-Type", "application/json")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}
}
