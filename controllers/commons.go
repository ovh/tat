package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/models"
	"github.com/ovh/tat/utils"
)

// PreCheckUser has to be called as a middleware on Gin Route.
// Check if username exists in database, return user if ok
func PreCheckUser(ctx *gin.Context) (models.User, error) {
	var user = models.User{}
	err := user.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		e := errors.New("Error while fetching user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": e})
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return user, e
	}
	return user, nil
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
