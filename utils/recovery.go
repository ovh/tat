package utils

import (
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// TatRecovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func TatRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				path := c.Request.URL.Path
				query := c.Request.URL.RawQuery
				username, _ := c.Get(TatHeaderUsername)
				trace := make([]byte, 4096)
				count := runtime.Stack(trace, true)
				log.Panicf("[TatRecovery] err:%s method:%s path:%s query:%s username:%s stacktrace of %d bytes:%s",
					err, c.Request.Method, path, query, username, count, trace)

				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
