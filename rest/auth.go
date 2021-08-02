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
