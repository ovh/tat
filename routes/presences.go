package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesPresences initialized routes for Presences Controller
func InitRoutesPresences(router *gin.Engine) {
	presencesCtrl := &controllers.PresencesController{}
	g := router.Group("/")
	g.Use(CheckPassword())
	{
		// List Presences
		g.GET("presences/*topic", presencesCtrl.List)
		// Add a presence and get list
		g.POST("presenceget/*topic", presencesCtrl.CreateAndGet)
		// delete a presence
		g.DELETE("presences/*topic", presencesCtrl.Delete)
	}
	admin := router.Group("/presencesadmin")
	admin.Use(CheckPassword(), CheckAdmin())
	{
		admin.GET("/checkall", presencesCtrl.CheckAllPresences)
	}
}
