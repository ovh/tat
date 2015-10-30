package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesTopics initialized routes for Topics Controller
func InitRoutesTopics(router *gin.Engine) {
	topicsCtrl := &controllers.TopicsController{}

	g := router.Group("/")
	g.Use(CheckPassword())
	{
		g.GET("/topics", topicsCtrl.List)
		g.POST("/topic", topicsCtrl.Create)
		g.DELETE("/topic/*topic", topicsCtrl.Delete)
		g.GET("/topic/*topic", topicsCtrl.OneTopic)

		g.PUT("/topic/add/parameter", topicsCtrl.AddParameter)
		g.PUT("/topic/remove/parameter", topicsCtrl.RemoveParameter)

		g.PUT("/topic/add/rouser", topicsCtrl.AddRoUser)
		g.PUT("/topic/remove/rouser", topicsCtrl.RemoveRoUser)
		g.PUT("/topic/add/rwuser", topicsCtrl.AddRwUser)
		g.PUT("/topic/remove/rwuser", topicsCtrl.RemoveRwUser)
		g.PUT("/topic/add/adminuser", topicsCtrl.AddAdminUser)
		g.PUT("/topic/remove/adminuser", topicsCtrl.RemoveAdminUser)

		g.PUT("/topic/add/rogroup", topicsCtrl.AddRoGroup)
		g.PUT("/topic/remove/rogroup", topicsCtrl.RemoveRoGroup)
		g.PUT("/topic/add/rwgroup", topicsCtrl.AddRwGroup)
		g.PUT("/topic/remove/rwgroup", topicsCtrl.RemoveRwGroup)
		g.PUT("/topic/add/admingroup", topicsCtrl.AddAdminGroup)
		g.PUT("/topic/remove/admingroup", topicsCtrl.RemoveAdminGroup)
		g.PUT("/topic/param", topicsCtrl.SetParam)
	}
}
