package sdk

import (
	datautils "github.com/soumitsalman/data-utils"
)

func (collector *RedditCollector) IsUserAuthenticated(userid string) (bool, string) {
	user := collector.GetCollectionAccount(userid)
	if user != nil {
		return true, ""
	}
	return false, GetRedditAuthorizationUrl(userid, collector.config.RedditClientConfig)
}

func (collector *RedditCollector) AddCollectionAccount(user RedditUser) {
	collector.authenticated_users = append(collector.authenticated_users, user)
}

func (collector *RedditCollector) GetCollectionAccount(userid string) *RedditUser {
	index := datautils.IndexAny(collector.authenticated_users, func(item *RedditUser) bool { return item.UserId == userid })
	if index >= 0 {
		return &collector.authenticated_users[index]
	}
	return nil
}
