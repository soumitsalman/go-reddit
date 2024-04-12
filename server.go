package main

import (
	"fmt"
	"os"

	"github.com/soumitsalman/goreddit/api"

	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type appAuthorizationParams struct {
	UserId string `form:"state"`
	Code   string `form:"code"`
	Error  string `form:"error"`
}

func collectHandler(ctx *gin.Context) {
	go api.CollectAndStoreAll()
	ctx.String(http.StatusOK, "collection started")
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

func redditOauthRedirectHandler(ctx *gin.Context) {
	var params appAuthorizationParams
	if ctx.BindQuery(&params) != nil || params.Error != "" {
		ctx.String(http.StatusUnauthorized, "Some authorization issue")
		return
	}

	client, err := api.NewOauthRedditClient(params.UserId, params.Code)
	if err != nil {
		ctx.String(http.StatusUnauthorized, "Some authorization issue")
		return
	}

	api.AddRedditUser(*client.User)
	ctx.String(http.StatusOK, "Authentication Succeeded for %s. Feel Free to Close the Window.", params.UserId)
}

func userAuthCheckHandler(ctx *gin.Context) {
	var params appAuthorizationParams
	if ctx.BindQuery(&params) == nil {
		ok, res := api.CheckAuthenticationStatus(params.UserId)
		if ok {
			ctx.String(http.StatusOK, res)
		} else {
			ctx.String(http.StatusNotFound, res)
		}
		return
	}
	ctx.Status(http.StatusBadRequest)
}

func authorizeHandler(ctx *gin.Context) {
	userid := "__BLANK__"
	var params appAuthorizationParams
	if ctx.BindQuery(&params) == nil {
		userid = params.UserId
	}

	data := []byte(fmt.Sprintf("<html><body><a href=\"%s\">Sign-in with Reddit</a></body></html>", api.GetRedditAuthorizationUrl(userid)))
	ctx.Data(http.StatusOK, "text/html", data)
}

func getInternalAuthToken() string {
	return os.Getenv("INTERNAL_AUTH_TOKEN")
}

func NewServer(r rate.Limit, b int) *gin.Engine {
	// runCollection()
	router := gin.Default()

	auth_group := router.Group("/")
	// authn and ratelimit middleware
	auth_group.Use(createRateLimitHandler(r, b))
	// routes
	auth_group.POST("/reddit/collect", collectHandler)
	auth_group.GET("/reddit/oauth-redirect", redditOauthRedirectHandler)
	auth_group.GET("/reddit/auth-status", userAuthCheckHandler)
	auth_group.GET("/reddit/authorize", authorizeHandler)

	return router
}

func main() {
	NewServer(2, 5).Run()
}
