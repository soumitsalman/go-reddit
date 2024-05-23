package examples

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	ds "github.com/soumitsalman/beansack/sdk"
	sdk "github.com/soumitsalman/go-reddit/collector"
)

func CollectAndStoreLocally() {
	config := sdk.NewCollectorConfig(localFileStore)
	sdk.NewCollector(config).Collect()
}

func localFileStore(contents []ds.Bean) {
	data, _ := json.MarshalIndent(contents, "", "\t")
	filename := fmt.Sprintf("test_REDDIT_%s", time.Now().Format("2006-01-02-15-04-05.json"))
	os.WriteFile(filename, data, 0644)
}

// TODO: delete all of these
// func redditOauthRedirectHandler(ctx *gin.Context) {
// 	var params appAuthorizationParams
// 	if ctx.BindQuery(&params) != nil || params.Error != "" {
// 		ctx.String(http.StatusUnauthorized, "Some authorization issue")
// 		return
// 	}

// 	client, err := sdk.NewOauthRedditClient(params.UserId, params.Code, config.RedditClientConfig)
// 	if err != nil {
// 		ctx.String(http.StatusUnauthorized, "Some authorization issue")
// 		return
// 	}

// 	sdk.AddRedditUser(*client.User)
// 	ctx.String(http.StatusOK, "Authentication Succeeded for %s. Feel Free to Close the Window.", params.UserId)
// }

// func userAuthCheckHandler(ctx *gin.Context) {
// 	var params appAuthorizationParams
// 	if ctx.BindQuery(&params) == nil {
// 		ok, res := sdk.IsUserAuthenticated(params.UserId)
// 		if ok {
// 			ctx.String(http.StatusOK, res)
// 		} else {
// 			ctx.String(http.StatusNotFound, res)
// 		}
// 		return
// 	}
// 	ctx.Status(http.StatusBadRequest)
// }

// func redditAuthorizeHandler(ctx *gin.Context) {
// 	userid := sdk.DEFAULT_USERID
// 	var params appAuthorizationParams
// 	if ctx.BindQuery(&params) == nil && params.UserId != "" {
// 		userid = params.UserId
// 	}

// 	data := []byte(fmt.Sprintf("<html><body><a href=\"%s\">Sign-in with Reddit</a></body></html>", sdk.GetRedditAuthorizationUrl(userid, config.RedditClientConfig)))
// 	ctx.Data(http.StatusOK, "text/html", data)
// }
