// middlewares/panic_recovery.go
package middlewares

import (
	"TeamTickBackend/gen"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
)

func recoverMiddleware(operationID string) {
	if r := recover(); r != nil {
		// 记录详细的堆栈信息，便于排查问题
		stackTrace := debug.Stack()
		fmt.Printf("Panic - operation: %s\nerror: %v\nstack: %s\n", 
			operationID, r, stackTrace)
		
		// 将panic转换为HTTP 500错误
		// 注意：这里并不直接返回，而是继续panic
		// 因为我们需要跳出当前函数，中断正常执行流
		// 外层的response, err := handler(ctx, request)需要捕获到这个错误
		panic(fmt.Errorf("internal server error: %v", r))
	}
}

func AuthRecoveryMiddleware() gen.AuthStrictMiddlewareFunc {
	return func(handler strictgin.StrictGinHandlerFunc, operationID string) strictgin.StrictGinHandlerFunc {
		return func(c*gin.Context,request interface{}) (interface{},error){
			defer recoverMiddleware(operationID)
			return handler(c,request)
		}
	}
}

func GroupRecoveryMiddleware() gen.AuthStrictMiddlewareFunc {
	return func(handler strictgin.StrictGinHandlerFunc, operationID string) strictgin.StrictGinHandlerFunc {
		return func(c*gin.Context,request interface{}) (interface{},error){
			defer recoverMiddleware(operationID)
			return handler(c,request)
		}
	}
}
func UserRecoveryMiddleware() gen.AuthStrictMiddlewareFunc {
	return func(handler strictgin.StrictGinHandlerFunc, operationID string) strictgin.StrictGinHandlerFunc {
		return func(c*gin.Context,request interface{}) (interface{},error){
			defer recoverMiddleware(operationID)
			return handler(c,request)
		}
	}
}
