package ws

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
			log.Printf("upgrade error: %v", err)
			return
		}

		client := NewClient(hub, conn, jwtSecret, sendMsg)
		go client.WritePump()
		go client.ReadPump()
	}
}