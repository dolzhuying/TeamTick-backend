package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
)

func RecoverMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 记录错误和堆栈信息
                stack := debug.Stack()
                logger.Errorf("Panic recovered: %v\n%s", err, stack)
                
                // 返回统一格式的JSON响应
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "code":    "1",
                    "message": "internal server error",
                })
            }
        }()
        c.Next()
    }
}