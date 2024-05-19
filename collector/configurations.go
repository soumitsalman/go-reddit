package sdk

import (
	"os"
	"time"

	ds "github.com/soumitsalman/beansack/sdk"
)

const (
	JSON_BODY        = "application/json"
	URL_ENCODED_BODY = "application/x-www-form-urlencoded"
)

const (
	// changing wait time since it takes while to publish all the messages
	MAX_WAIT_TIME = 10 * time.Minute
)

type RedditClientConfig struct {
	AppName     string
	AppId       string
	AppSecret   string
	RedirectUri string
	Scope       string
}

// type BeansackConfig struct {
// 	BeanSackUrl    string
// 	BeanSackAPIKey string
// 	UserAgent      string
// }

type CollectorConfig struct {
	MasterCollectorUsername string
	MasterCollectorPassword string
	RedditClientConfig
	store_func func(beans []ds.Bean)
}

const (
	DEFAULT_USERID = "__BLANK__"
	SCOPE          = "identity read mysubreddits"
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

// func getBeanUrl() string {
// 	return os.Getenv("BEANSACK_URL")
// }

// func getInternalAuthToken() string {
// 	return os.Getenv("INTERNAL_AUTH_TOKEN")
// }

func NewCollectorConfig(store_func func(beans []ds.Bean)) CollectorConfig {
	return CollectorConfig{
		// BeansackConfig: BeansackConfig{
		// 	BeanSackUrl:    getBeanUrl(),
		// 	BeanSackAPIKey: getInternalAuthToken(),
		// 	UserAgent:      getAppName(),
		// },
		MasterCollectorUsername: getMasterUsername(),
		MasterCollectorPassword: getMasterPassword(),
		RedditClientConfig: RedditClientConfig{
			AppName:     getAppName(),
			AppId:       getAppId(),
			AppSecret:   getAppSecret(),
			RedirectUri: getOauthRedirectUri(),
			Scope:       SCOPE,
		},
		store_func: store_func,
	}
}
