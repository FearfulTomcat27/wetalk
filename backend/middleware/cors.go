package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 返回 CORS 中间件，允许 localhost 和局域网 3000 端口访问
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 开发环境：允许 localhost 和局域网 3000 端口访问
			if origin == "http://localhost:3000" {
				return true
			}
			// 允许局域网内任何 IP 的 3000 端口访问
			return strings.HasPrefix(origin, "http://192.168.") && strings.HasSuffix(origin, ":3000")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}
