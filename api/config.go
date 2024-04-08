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

func GetAppName() string {
	return os.Getenv("REDDITOR_APP_NAME")
}

func GetOauthRedirectUri() string {
	return os.Getenv("REDDITOR_HOST") + "/reddit/oauth_redirect"
}

func GetAppId() string {
	return os.Getenv("REDDITOR_APP_ID")
}

func GetAppSecret() string {
	return os.Getenv("REDDITOR_APP_SECRET")
}

func getMasterUserName() string {
	return os.Getenv("REDDITOR_MASTER_USER_NAME")
}

func getMasterUserPw() string {
	return os.Getenv("REDDITOR_MASTER_USER_PW")
}

func getUserAgent() string {
	//Windows:My Reddit Bot:1.0 (by u/botdeveloper)
	return fmt.Sprintf("%v:%s:v0.1 (by u/%s)", runtime.GOOS, getMasterUserName(), GetAppName())
}

func getBeanUrl() string {
	return os.Getenv("BEANSACK_URL")
}
