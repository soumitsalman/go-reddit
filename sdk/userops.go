package sdk

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
var authenticated_users []RedditUser = make([]RedditUser, 0, 100)

func GetRedditUsers() []RedditUser {
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
