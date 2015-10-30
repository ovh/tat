package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
	"github.com/spf13/viper"
)

// InitRoutesSockets initialized routes for Sockets Controller
func InitRoutesSockets(router *gin.Engine) {
	socketsCtrl := &controllers.SocketsController{}

	if viper.GetBool("websocket_enabled") {
		router.GET("/socket/ws", socketsCtrl.WS)

		admin := router.Group("/sockets")
		admin.Use(CheckPassword(), CheckAdmin())
		{
			admin.GET("/dump", socketsCtrl.Dump)
		}
	}

}
