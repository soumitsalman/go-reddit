package main

import (
	"os"

	"github.com/soumitsalman/goreddit/sdk"
)

const (
	_DEFAULT_USERID = "__BLANK__"
	_SCOPE          = "identity read mysubreddits"
)

func getAppName() string {
	return os.Getenv("REDDITOR_APP_NAME")
}

func getOauthRedirectUri() string {
	return os.Getenv("REDDITOR_OAUTH_REDIRECT_URI")
}

func getAppId() string {
	return os.Getenv("REDDITOR_APP_ID")
}

func getAppSecret() string {
	return os.Getenv("REDDITOR_APP_SECRET")
}

func getMasterUsername() string {
	return os.Getenv("REDDITOR_MASTER_USER_NAME")
}

func getMasterPassword() string {
	return os.Getenv("REDDITOR_MASTER_USER_PW")
}

func getBeanUrl() string {
	return os.Getenv("BEANSACK_URL")
}

func getInternalAuthToken() string {
	return os.Getenv("INTERNAL_AUTH_TOKEN")
}

func getCollectorConfig() sdk.RedditCollectorConfig {
	return sdk.RedditCollectorConfig{
		BeansackConfig: sdk.BeansackConfig{
			BeanSackUrl:    getBeanUrl(),
			BeanSackAPIKey: getInternalAuthToken(),
			UserAgent:      getAppName(),
		},
		MasterCollectorUsername: getMasterUsername(),
		MasterCollectorPassword: getMasterPassword(),
		RedditClientConfig: sdk.RedditClientConfig{
			AppName:     getAppName(),
			AppId:       getAppId(),
			AppSecret:   getAppSecret(),
			RedirectUri: getOauthRedirectUri(),
			Scope:       _SCOPE,
		},
	}
}
