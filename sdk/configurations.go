package sdk

import (
	"time"
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
