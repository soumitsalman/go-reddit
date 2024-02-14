package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func collectHandler(ctx *gin.Context) {
	go runCollections()
	ctx.JSON(http.StatusOK, gin.H{"message": "collection started"})
}

func runCollections() {
	for _, user := range GetRedditUsers() {
		NewCollectorClient(&user).CollectItems()
	}
}

func authenticationHandler(ctx *gin.Context) {
	// log.Println(ctx.GetHeader("X-API-Key"), getInternalAuthToken())
	if ctx.GetHeader("X-API-Key") == getInternalAuthToken() {
		ctx.Next()
	} else {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
}

func createRateLimitHandler(r rate.Limit, b int) gin.HandlerFunc {
	rate_limiter := rate.NewLimiter(r, b)
	return func(ctx *gin.Context) {
		if rate_limiter.Allow() {
			ctx.Next()
		} else {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}

func NewServer(r rate.Limit, b int) *gin.Engine {
	// runCollection()
	router := gin.Default()

	auth_group := router.Group("/")
	// authn and ratelimit middleware
	auth_group.Use(createRateLimitHandler(r, b), authenticationHandler)
	// routes
	auth_group.GET("/collect", collectHandler)

	return router
}
