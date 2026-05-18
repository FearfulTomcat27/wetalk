package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"wetalk/common"
	"wetalk/types"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.AppError(c, types.ErrMissingToken)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.AppError(c, types.ErrInvalidTokenFormat)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			common.AppError(c, types.ErrInvalidClaims)
			return
		}

		// 将用户 ID 存入上下文
		userID, _ := claims["user_id"].(float64)
		c.Set("user_id", int64(userID))
		c.Next()
	}
}
