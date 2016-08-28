package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesMessages initialized routes for Messages Controller
func InitRoutesMessages(router *gin.Engine) {
	messagesCtrl := &controllers.MessagesController{}

	g := router.Group("/messages")
	g.Use(CheckPassword())
	{
		g.GET("/*topic", messagesCtrl.List)
		g.DELETE("/nocascade/*topic", messagesCtrl.DeleteBulk)
		g.DELETE("/cascade/*topic", messagesCtrl.DeleteBulkCascade)
		g.DELETE("/cascadeforce/*topic", messagesCtrl.DeleteBulkCascadeForce)
	}

	r := router.Group("/read")
	r.Use()
	{
		r.GET("/*topic", messagesCtrl.List)
	}

	gm := router.Group("/message")
	gm.Use(CheckPassword())
	{
		//Create a message, a reply
		gm.POST("/*topic", messagesCtrl.Create)

		// Like, Unlike, Label, Unlabel a message, mark as task, voteup, votedown, unvoteup, unvotedown
		gm.PUT("/*topic", messagesCtrl.Update)

		// Delete a message
		gm.DELETE("/nocascade/:idMessage/*topic", messagesCtrl.Delete)

		// Delete a message and its replies
		gm.DELETE("/cascade/:idMessage/*topic", messagesCtrl.DeleteCascade)

		// Delete a message and its replies, event if it's in a Tasks Topic of one user
		gm.DELETE("/cascadeforce/:idMessage/*topic", messagesCtrl.DeleteCascadeForce)
	}
}
