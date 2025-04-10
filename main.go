package main

import (
	"TeamTickBackend/app"
	"TeamTickBackend/router"
)

func main() {
	container := app.NewAppContainer()
	router := router.SetupRouter(container)
	router.Run(":8080")
}