package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

func ginHTTPHandler(handler httpHandlerFunc, pathParams ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, name := range pathParams {
			c.Request.SetPathValue(name, c.Param(name))
		}
		handler(c.Writer, c.Request)
	}
}

func appHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Robot-Center", "app-server")
		c.Next()
	}
}
