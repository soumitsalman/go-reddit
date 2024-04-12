package sdk

import (
	"time"
)

const (
	JSON_BODY        = "application/json"
	URL_ENCODED_BODY = "application/x-www-form-urlencoded"
)

const (
	MAX_WAIT_TIME = 30 * time.Second
)

type RedditClientConfig struct {
	AppName     string
	AppId       string
	AppSecret   string
	RedirectUri string
	Scope       string
}

type BeansackConfig struct {
	BeanSackUrl    string
	BeanSackAPIKey string
	UserAgent      string
}

type RedditCollectorConfig struct {
	MasterCollectorUsername string
	MasterCollectorPassword string
	RedditClientConfig
	BeansackConfig
}

// func (config *RedditClientConfig) GetUserAgent() string {
// 	return fmt.Sprintf("%v:%s:v0.1", runtime.GOOS, config.AppName)
// }
