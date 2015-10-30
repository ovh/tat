package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesGroups initialized routes for Groups Controller
func InitRoutesGroups(router *gin.Engine) {

	groupsCtrl := &controllers.GroupsController{}

	g := router.Group("/")
	g.Use(CheckPassword())
	{
		g.GET("/groups", groupsCtrl.List)

		g.PUT("/group/add/user", groupsCtrl.AddUser)
		g.PUT("/group/remove/user", groupsCtrl.RemoveUser)
		g.PUT("/group/add/adminuser", groupsCtrl.AddAdminUser)
		g.PUT("/group/remove/adminuser", groupsCtrl.RemoveAdminUser)

		admin := router.Group("/group")
		admin.Use(CheckPassword(), CheckAdmin())
		{
			admin.POST("", groupsCtrl.Create)
		}
	}
}
