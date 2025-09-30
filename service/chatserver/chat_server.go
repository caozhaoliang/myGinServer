package chatserver

import (
	"fmt"
	"log"
	"myGinServer/config"
	"myGinServer/internal/cache"
	"myGinServer/models/chat_msg"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatServer struct {
	cache cache.StoreCache
}

const (
	ChatChannelKey = "CHAT_CHANNEL_KEY"
)

func NewChatServer(cacheConfig *config.CacheConfig) *ChatServer {
	storeCache := cache.NewRedisCache(cacheConfig)
	return &ChatServer{cache: storeCache}
}

var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ch *ChatServer) SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(wsConn *websocket.Conn) {
		err = wsConn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	ch.MsgHandler(ws, c)
}
func (ch *ChatServer) MsgHandler(ws *websocket.Conn, c *gin.Context) {
	go func() {
		for {
			msg, err := ch.cache.Subscribe(c, ChatChannelKey)
			if err != nil {
				fmt.Println(err)
				return
			}
			now := time.Now().Format("2006-01-02 15:04:05")
			fmtMsg := fmt.Sprintf("[ws][%s]:%s", now, msg)
			err = ws.WriteMessage(1, []byte(fmtMsg))
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("读取客户端消息失败: %v", err)
			break
		}

		// 将消息发布到Redis频道
		err = ch.cache.Publish(c, ChatChannelKey, string(msg))
		if err != nil {
			log.Printf("发布消息到Redis失败: %v", err)
		}
	}
}

func (ch *ChatServer) SendUserMsg(c *gin.Context) {
	chat_msg.Chat(c.Writer, c.Request)
}
