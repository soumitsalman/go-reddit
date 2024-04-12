package api

import (
	datautils "github.com/soumitsalman/data-utils"
)

const (
	REDDIT_URL    = "http://www.reddit.com"
	REDDIT_SOURCE = "REDDIT"
)

const (
	_MASTER_COLLECTOR = "__DEFAULT_MASTER_COLLECTOR__"
)

// LOGGED-IN USERS RELATED FUNCTIONS
var authenticated_users []RedditUser

func GetRedditUsers() []RedditUser {
	// initialize with default master
	if authenticated_users == nil {
		authenticated_users = []RedditUser{
			{
				UserId:   _MASTER_COLLECTOR,
				Username: getMasterUserName(),
				Password: getMasterUserPw(),
			},
		}
	}
	return authenticated_users
}

func AddRedditUser(user RedditUser) {
	authenticated_users = append(authenticated_users, user)
}

func GetRedditUser(userid string) *RedditUser {
	index := datautils.IndexAny(authenticated_users, func(item *RedditUser) bool { return item.UserId == userid })
	if index >= 0 {
		return &authenticated_users[index]
	}
	return nil
}

func CheckAuthenticationStatus(userid string) (bool, string) {
	user := GetRedditUser(userid)
	if user != nil {
		if client, err := NewRedditClient(user); err == nil {
			return true, client.User.AccessToken
		}
	}
	return false, GetRedditAuthorizationUrl(userid)
}
