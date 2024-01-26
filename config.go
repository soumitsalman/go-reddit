package main

import (
	"os"
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

func getLocalUserName() string {
	return os.Getenv("_REDDIT_LOCAL_USER_NAME")
}

func getLocalUserPw() string {
	return os.Getenv("_REDDIT_LOCAL_USER_PW")
}
