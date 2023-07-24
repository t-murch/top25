package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func Run() {
	getRoutes()
	router.Run(":8080")
}

func getRoutes() {
	v1 := router.Group("/v1")
	v1.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Top 25 API up and running. ")
	})
	addSpotifyRoutes(v1)
}
