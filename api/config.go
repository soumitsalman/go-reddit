package api

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	JSON_BODY        = "application/json"
	URL_ENCODED_BODY = "application/x-www-form-urlencoded"
)

const (
	MAX_WAIT_TIME = 30 * time.Second
)

func getAppName() string {
	return os.Getenv("REDDIT_APP_NAME")
}

func getAppDescription() string {
	return os.Getenv("REDDIT_APP_DESCRIPTION")
}

func getAboutUrl() string {
	return os.Getenv("REDDIT_ABOUT_URL")
}

func getRedirectUri() string {
	return os.Getenv("REDDIT_REDIRECT_URI")
}

func getAppId() string {
	return os.Getenv("REDDIT_APP_ID")
}

func getAppSecret() string {
	return os.Getenv("REDDIT_APP_SECRET")
}

func getUserAgent() string {
	//Windows:My Reddit Bot:1.0 (by u/botdeveloper)
	return fmt.Sprintf("%v:%v:v0.1 (by u/randomizer_000)", runtime.GOOS, getAppName())
}

// func getLocalUserName() string {
// 	return os.Getenv("_REDDIT_LOCAL_USER_NAME")
// }

// func getLocalUserPw() string {
// 	return os.Getenv("_REDDIT_LOCAL_USER_PW")
// }

func getInternalAuthToken() string {
	return os.Getenv("INTERNAL_AUTH_TOKEN")
}

func getMediaStoreUrl() string {
	return os.Getenv("MEDIA_STORE_URL")
}
