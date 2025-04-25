package middlewares

import (
	"TeamTickBackend/pkg"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 认证中间件，错误处理日志待完善
func AuthMiddleware(jwtToken pkg.JwtHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "1",
				"message": "missing authorization",
			})
			return
		}

		payload, err := jwtToken.ParseJWTToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "1",
				"message": "invalid token",
			})
			return
		}

		c.Set("username", payload.Username)
		c.Set("userID", payload.UserID)
		c.Set("authenticated", true)
		c.Set("auth_time", time.Now().Unix())

		// 同时存储到请求上下文（handlers层接受的是标准库Context）
		ctx := context.WithValue(c.Request.Context(), "userID", payload.UserID)
		ctx = context.WithValue(ctx, "username", payload.Username)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
