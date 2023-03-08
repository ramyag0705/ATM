package main

import (
	"account/configs"
	"account/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//run database
	configs.ConnectDB()

	//routes
	routes.UserRoute(r)

	r.Run("localhost:6000")
}

