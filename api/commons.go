package main

import (
	"errors"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat"
	userDB "github.com/ovh/tat/api/user"
)

// PreCheckUser has to be called as a middleware on Gin Route.
// Check if username exists in database, return user if ok
func PreCheckUser(ctx *gin.Context) (tat.User, error) {
	var tatUser = tat.User{}
	found, err := userDB.FindByUsername(&tatUser, getCtxUsername(ctx))
	var e error
	if !found {
		e = errors.New("User unknown")
	} else if err != nil {
		e = errors.New("Error while fetching user")
	}
	if e != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": e})
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return tatUser, e
	}
	return tatUser, nil
}

// GetParam returns the value of a parameter in Url.
// Example : http://host:port/:paramName
func GetParam(ctx *gin.Context, paramName string) (string, error) {
	value, found := ctx.Params.Get(paramName)
	if !found {
		s := paramName + " in url does not exist"
		ctx.JSON(http.StatusBadRequest, gin.H{"error": s})
		return "", errors.New(s)
	}
	return value, nil
}

// AbortWithReturnError abort gin context and return JSON to user with error details
func AbortWithReturnError(ctx *gin.Context, statusHTTP int, err error) {
	ctx.JSON(statusHTTP, gin.H{"error:": err.Error()})
	ctx.Abort()
}

// tatRecovery is a middleware that recovers from any panics and writes a 500 if there was one.
func tatRecovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			path := c.Request.URL.Path
			query := c.Request.URL.RawQuery
			username, _ := c.Get(tat.TatHeaderUsername)
			trace := make([]byte, 4096)
			count := runtime.Stack(trace, true)
			log.Panicf("[tatRecovery] err:%s method:%s path:%s query:%s username:%s stacktrace of %d bytes:%s",
				err, c.Request.Method, path, query, username, count, trace)

			c.AbortWithStatus(500)
		}
	}()
	c.Next()

}

// ginrus returns a gin.HandlerFunc (middleware) that logs requests using logrus.
//
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func ginrus(logger *logrus.Logger, timeFormat string, utc bool) gin.HandlerFunc {
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

		username, _ := c.Get(tat.TatHeaderUsername)
		tatReferer, _ := c.Get(tat.TatHeaderXTatRefererLower)

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
