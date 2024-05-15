package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/soumitsalman/goreddit/sdk"

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
	go sdk.CollectAndStoreAll()
	ctx.String(http.StatusOK, "collection started")
}

func serverAuthenticationHandler(ctx *gin.Context) {
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

	client, err := sdk.NewOauthRedditClient(params.UserId, params.Code, config.RedditClientConfig)
	if err != nil {
		ctx.String(http.StatusUnauthorized, "Some authorization issue")
		return
	}

	sdk.AddRedditUser(*client.User)
	ctx.String(http.StatusOK, "Authentication Succeeded for %s. Feel Free to Close the Window.", params.UserId)
}

func userAuthCheckHandler(ctx *gin.Context) {
	var params appAuthorizationParams
	if ctx.BindQuery(&params) == nil {
		ok, res := sdk.IsUserAuthenticated(params.UserId)
		if ok {
			ctx.String(http.StatusOK, res)
		} else {
			ctx.String(http.StatusNotFound, res)
		}
		return
	}
	ctx.Status(http.StatusBadRequest)
}

func redditAuthorizeHandler(ctx *gin.Context) {
	userid := _DEFAULT_USERID
	var params appAuthorizationParams
	if ctx.BindQuery(&params) == nil && params.UserId != "" {
		userid = params.UserId
	}

	data := []byte(fmt.Sprintf("<html><body><a href=\"%s\">Sign-in with Reddit</a></body></html>", sdk.GetRedditAuthorizationUrl(userid, config.RedditClientConfig)))
	ctx.Data(http.StatusOK, "text/html", data)
}

var config sdk.RedditCollectorConfig

func NewServer(r rate.Limit, b int) *gin.Engine {
	config = getCollectorConfig()
	sdk.Initialize(config)

	router := gin.Default()

	// Does not need auth since they should be publicly accessible
	noauth_group := router.Group("/reddit")
	noauth_group.Use(createRateLimitHandler(r, b))
	noauth_group.GET("/authorize", redditAuthorizeHandler)
	noauth_group.GET("/oauth-redirect", redditOauthRedirectHandler)

	// NEEDS SERVICES TO SERVICE AUTH
	auth_group := router.Group("/reddit")
	auth_group.Use(createRateLimitHandler(r, b), serverAuthenticationHandler)

	auth_group.POST("/collect", collectHandler)
	auth_group.GET("/auth-status", userAuthCheckHandler)

	return router
}

func main() {
	godotenv.Load()
	NewServer(2, 5).Run()
}
