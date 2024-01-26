package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func collectHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, RunBulkCollect())
}

func main() {
	router := gin.Default()
	router.GET("/collect", collectHandler)
	router.Run()
	// RunBulkCollect()
}
