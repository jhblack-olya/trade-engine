/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/gae4/trade-engine/conf"
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
func healthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, "Ok")
		return

	}
}
