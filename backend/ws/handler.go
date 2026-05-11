package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"wetalk/common/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler WebSocket 连接处理器
func WSHandler(hub *Hub, jwtSecret string, sendMsg SendMessageFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Module("ws").Error("upgrade error", "err", err)
			return
		}

		client := NewClient(hub, conn, jwtSecret, sendMsg)
		go client.WritePump()
		go client.ReadPump()
	}
}
