package router

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	"TeamTickBackend/handlers"
	"TeamTickBackend/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter(container *app.AppContainer) *gin.Engine {
	router := gin.Default()

	authHandler := handlers.NewAuthHandler(container)
	gen.RegisterAuthHandlers(router, authHandler)

	userHandler := handlers.NewUserHandler(container)
	userRouter := router.Group("")
	userRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterUsersHandlers(userRouter, userHandler)

	groupsHandler := handlers.NewGroupsHandler(container)
	groupsRouter := router.Group("")
	groupsRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterGroupsHandlers(groupsRouter, groupsHandler)

	auditRequestHandler := handlers.NewAuditRequestHandler(container)
	auditRequestRouter := router.Group("")
	auditRequestRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterAuditRequestHandlers(auditRequestRouter, auditRequestHandler)

	return router
}
