package controller

import (
	"myGinServer/config"
	"myGinServer/service/chatserver"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	chatServer *chatserver.ChatServer
}

func NewChatController(config *config.Config) *ChatController {
	chatServer := chatserver.NewChatServer(&config.Cache)
	return &ChatController{chatServer: chatServer}
}

func (ch *ChatController) SendMsg(c *gin.Context) {
	ch.chatServer.SendMsg(c)
}

func (ch *ChatController) SendUserMsg(c *gin.Context) {
	ch.chatServer.SendUserMsg(c)
}
