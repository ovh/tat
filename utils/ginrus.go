package utils

//
// inspired from https://github.com/gin-gonic/contrib/blob/master/ginrus/ginrus.go
//

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// Ginrus returns a gin.HandlerFunc (middleware) that logs requests using logrus.
//
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func Ginrus(logger *logrus.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		username, _ := c.Get(TatHeaderUsername)
		tatReferer, _ := c.Get(TatHeaderXTatRefererLower)

		entry := logger.WithFields(logrus.Fields{
			"status":      c.Writer.Status(),
			"method":      c.Request.Method,
			"path":        path,
			"query":       query,
			"ip":          c.ClientIP(),
			"latency":     latency,
			"user-agent":  c.Request.UserAgent(),
			"time":        end.Format(timeFormat),
			"tatusername": username,
			"tatfrom":     tatReferer,
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else if c.Writer.Status() >= 400 {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
