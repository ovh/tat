package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesUsers initialized routes for Users Controller
func InitRoutesUsers(router *gin.Engine) {
	usersCtrl := &controllers.UsersController{}

	gs := router.Group("/users")
	gs.Use(CheckPassword())
	{
		gs.GET("", usersCtrl.List)
	}
	g := router.Group("/user")
	g.Use(CheckPassword())
	{
		g.GET("/me", usersCtrl.Me)
		g.GET("/me/contacts/:sinceSeconds", usersCtrl.Contacts)
		g.POST("/me/contacts/:username", usersCtrl.AddContact)
		g.DELETE("/me/contacts/:username", usersCtrl.RemoveContact)
		g.POST("/me/topics/*topic", usersCtrl.AddFavoriteTopic)
		g.DELETE("/me/topics/*topic", usersCtrl.RemoveFavoriteTopic)
		g.POST("/me/tags/:tag", usersCtrl.AddFavoriteTag)
		g.DELETE("/me/tags/:tag", usersCtrl.RemoveFavoriteTag)

		g.POST("/me/enable/notifications/topics/*topic", usersCtrl.EnableNotificationsTopic)
		g.POST("/me/disable/notifications/topics/*topic", usersCtrl.DisableNotificationsTopic)
	}

	admin := router.Group("/user")
	admin.Use(CheckPassword(), CheckAdmin())
	{
		admin.PUT("/convert", usersCtrl.Convert)
		admin.PUT("/archive", usersCtrl.Archive)
		admin.PUT("/rename", usersCtrl.Rename)
		admin.PUT("/update", usersCtrl.Update)
		admin.PUT("/setadmin", usersCtrl.SetAdmin)
		admin.PUT("/resetsystem", usersCtrl.ResetSystemUser)
	}

	router.GET("/user/verify/:username/:tokenVerify", usersCtrl.Verify)
	router.POST("/user/reset", usersCtrl.Reset)
	router.POST("/user", usersCtrl.Create)
}
