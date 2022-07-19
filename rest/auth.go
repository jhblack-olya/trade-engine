/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/siddontang/go/log"
	"gitlab.com/gae4/trade-engine/conf"
	"gitlab.com/gae4/trade-engine/models"
)

func checkAPIkey() gin.HandlerFunc {
	return func(c *gin.Context) {

		apiKey := c.Request.Header.Get("Authorization")
		if len(apiKey) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, newMessageVo(errors.New("token not found")))
			return
		}
		gbeConfig := conf.GetConfig()
		if gbeConfig.ApiKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, newMessageVo(errors.New("Bad token")))
			return
		}
		c.Next()
	}
}

type health struct {
	Redis string `json:"redis"`
	Mysql string `json:"mysql"`
	Kafka string `json:"kafka"`
}

func healthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			//delete(models.CommonError, "kafka")
			delete(models.CommonError, "redis")
			delete(models.CommonError, "mysql")
			delete(models.CommonError, "kafka")

		}()
		health := health{}
		health.Kafka = ""
		health.Mysql = ""
		health.Redis = ""
		message := ""
		status := http.StatusOK
		if val, ok := models.CommonError["redis"]; ok {
			health.Redis = "[redis]: " + val + " "
			log.Error("Redis health declined ", health.Redis)
			status = http.StatusBadRequest
		}
		if val, ok := models.CommonError["mysql"]; ok {
			health.Mysql = "[DB]: " + val + " "
			log.Error("DB health declined ", health.Mysql)
			status = http.StatusBadRequest
		}
		if val, ok := models.CommonError["kafka"]; ok {
			health.Kafka = "[kafka]: " + val + " "
			log.Error("Kafka health declined ", health.Kafka)
			status = http.StatusBadRequest
		}
		if health.Kafka != "" || health.Mysql != "" || health.Redis != "" {
			message = health.Kafka + health.Mysql + health.Redis
		} else {
			message = "ok"
		}
		c.JSON(status, message)
		return

	}
}
