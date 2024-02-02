package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soumitsr/goreddit/reddit"
)

func collectHandler(ctx *gin.Context) {
	go runCollection()
	ctx.JSON(http.StatusOK, gin.H{"message": "collection started"})
}

func runCollection() {
	start_time := time.Now()
	contents, engagements := reddit.CollectAllUserItems()
	reddit.NewContents(contents)
	reddit.NewEngagements(engagements)
	log.Println("Execution Time: ", time.Now().Sub(start_time))
}

func main() {
	runCollection()
	// router := gin.Default()
	// router.GET("/collect", collectHandler)
	// router.Run()
}
