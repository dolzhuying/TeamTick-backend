package main

import (
	"github.com/gin-gonic/gin"

	"TeamTickBackend/handlers"
)

func main() {
	r := gin.Default()

	// 注册路由

	// 启动服务器
	r.Run(":8080")
}
