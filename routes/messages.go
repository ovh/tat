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

		// Delete a message and its replies
		g.DELETE("/cascade/:idMessage", messagesCtrl.DeleteCascade)

		// Delete a message and its replies, event if it's in a Tasks Topic of one user
		g.DELETE("/cascadeforce/:idMessage", messagesCtrl.DeleteCascadeForce)
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
		gm.DELETE("/:idMessage", messagesCtrl.Delete)
	}

	g.Use(CheckPassword(), CheckAdmin())
	{
		g.POST("/countMessagesEmptyTopic", messagesCtrl.CountEmptyTopic)
	}

}
