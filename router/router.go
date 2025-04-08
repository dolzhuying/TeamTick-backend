package router

import (
	"TeamTickBackend/gen"
	"TeamTickBackend/handlers"
	"TeamTickBackend/middlewares"
	"TeamTickBackend/pkg"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	
	authHandler := handlers.NewAuthHandler()
	gen.RegisterAuthHandlers(router, authHandler)

	userHandler := handlers.NewUserHandler()
	userRouter := router.Group("")
	userRouter.Use(middlewares.AuthMiddleware(pkg.JwtTokenInstance))//实例化具体位置待考虑
	gen.RegisterUsersHandlers(userRouter, userHandler)

	groupsHandler := handlers.NewGroupsHandler()
	groupsRouter := router.Group("")
	groupsRouter.Use(middlewares.AuthMiddleware(pkg.JwtTokenInstance))//实例化具体位置待考虑
	gen.RegisterGroupsHandlers(groupsRouter, groupsHandler)
	
	return router
}