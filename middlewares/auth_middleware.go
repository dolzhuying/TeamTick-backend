package middlewares

import (
	"TeamTickBackend/pkg"
	"TeamTickBackend/pkg/logger"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// 认证中间件，错误处理日志待完善
func AuthMiddleware(jwtToken pkg.JwtHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			logger.Error("认证失败：缺少Authorization头",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
				zap.String("headers", strings.Join(c.Request.Header["Authorization"], ",")),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "1",
				"message": "missing authorization",
			})
			return
		}

		// 检查token格式
		if !strings.HasPrefix(strings.ToUpper(token), "BEARER ") {
			logger.Error("认证失败：token格式错误",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
				zap.String("token_format", "缺少BEARER前缀"),
				zap.String("token", token),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "1",
				"message": "invalid token format: missing BEARER prefix",
			})
			return
		}

		payload, err := jwtToken.ParseJWTToken(token)
		if err != nil {
			// 解析具体的错误类型
			var errorType string
			switch {
			case errors.Is(err, jwt.ErrTokenMalformed):
				errorType = "token格式错误"
			case errors.Is(err, jwt.ErrTokenUnverifiable):
				errorType = "token无法验证"
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				errorType = "token签名无效"
			case errors.Is(err, jwt.ErrTokenExpired):
				errorType = "token已过期"
			case errors.Is(err, jwt.ErrTokenNotValidYet):
				errorType = "token尚未生效"
			default:
				errorType = "未知错误"
			}

			logger.Error("认证失败：JWT解析错误",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
				zap.String("error_type", errorType),
				zap.String("token_length", string(len(token))),
				zap.String("token_prefix", token[:min(20, len(token))]+"..."),
				zap.Error(err),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "1",
				"message": "invalid token: " + errorType,
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
