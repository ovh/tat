package controllers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// SocketsController contains all methods about messages manipulation
type SocketsController struct{}

// Dump dump ws current var
func (*SocketsController) Dump(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, models.SocketsDump())
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WS handle websocket
func (*SocketsController) WS(ctx *gin.Context) {
	wshandler(ctx.Writer, ctx.Request)
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to set websocket upgrade: %+v", err)
		return
	}

	socket := &models.Socket{Connection: conn}
	socket.Open()

}
