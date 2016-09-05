package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ovh/tat"
	"github.com/spf13/viper"
)

// SystemController contains all methods about version
type SystemController struct{}

//GetVersion returns version of tat
func (*SystemController) GetVersion(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"version": tat.Version})
}

//GetCapabilites returns version of tat
func (*SystemController) GetCapabilites(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"websocket_enabled":   viper.GetBool("websocket_enabled"),
		"username_from_email": viper.GetBool("username_from_email"),
	})
}
