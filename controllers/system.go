package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// VERSION is version of Tat.
const VERSION = "1.2.0"

// SystemController contains all methods about version
type SystemController struct{}

//GetVersion returns version of tat
func (*SystemController) GetVersion(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"version": VERSION})
}

//GetCapabilites returns version of tat
func (*SystemController) GetCapabilites(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"websocket_enabled":   viper.GetBool("websocket_enabled"),
		"username_from_email": viper.GetBool("username_from_email"),
	})
}
