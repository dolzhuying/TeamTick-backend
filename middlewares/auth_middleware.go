package middlewares

import (
	"TeamTickBackend/pkg"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 认证中间件，错误处理日志待完善
func AuthMiddleware(jwtToken *pkg.JwtToken) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.Status(http.StatusUnauthorized)
			ctx.Error(errors.New("missing authorization"))
			ctx.Abort()
			return
		}
		payload, err := jwtToken.ParseJWTToken(token)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			ctx.Error(err)
			ctx.Abort()
			return
		}
		ctx.Set("username", payload.Username)
		ctx.Set("userID", payload.UserID)
		ctx.Set("authenticated", true)
		ctx.Set("auth_time", time.Now().Unix())
		ctx.Next()
	}
}
