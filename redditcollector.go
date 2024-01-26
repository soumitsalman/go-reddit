package main

import (
	"log"
	"time"
)

type RedditUser struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	AuthToken string `json:"auth_token"`
}

func RunBulkCollect() map[string]any {

	start_time := time.Now()
	var result map[string]any = make(map[string]any)
	users := GetUsers()
	for _, user := range users {
		var items, engagements, name = CollectRedditItems(&user)
		result[name] = map[string]any{
			"name":                        name,
			"total_items":                 len(items),
			"total_items_with_engagement": len(engagements),
		}
		log.Printf("Collected %d items for %s. %d has engagements", len(items), name, len(engagements))
	}
	result["execution_time"] = time.Now().Sub(start_time).String()
	return result
}

func CollectRedditItems(user *RedditUser) ([]RedditItem, []string, string) {
	var collected_items []RedditItem
	var user_engagements []string
	var me_data RedditItem

	// inline collect function to add to collections
	collect := func(items []RedditItem) {
		// add to the total collection
		collected_items = append(collected_items, items...)

		// add to the items where the user is already a subscriber or is an author
		for _, item := range items {
			if item.UserIsContributor || item.UserIsSubscriber || item.UserIsModerator || me_data.Name == item.Author {
				user_engagements = append(user_engagements, item.Name)
			}
		}
	}

	// instantiate client for data collection
	var rdclient = GetRedditClient(user)

	// get user data to match author name
	me_data, _ = rdclient.Me()

	// collect subreddits the user is already subscribing to
	var subreddits, _ = rdclient.Subreddits()
	collect(subreddits)
	log.Println(len(subreddits), "subreddits found")

	for _, sr := range subreddits {
		// TODO: check if this item related data has already been explored for this session
		// load subreddits similar to this subreddit
		if similar, err := rdclient.SimilarSubreddits(&sr); err != nil {
			log.Println(err)
		} else {
			collect(similar)
			log.Println(len(similar), "similar subreddits found for /r/", sr.DisplayName)
		}

		// load the hot posts in this subreddit
		if posts, err := rdclient.Posts(&sr, HOT); err != nil {
			log.Println(err)
		} else {
			collect(posts)
			log.Println(len(posts), "hot posts found for /r/", sr.DisplayName)

			// retrieve comments from this post
			// TODO: enable comment collection
			// for _, p := range posts {
			// 	var comments, _ = rdclient.RetrieveComments(&p)
			// 	collect(comments)
			// 	log.Println(len(comments), "comments found for", p.Name)
			// }
		}
	}
	return collected_items, user_engagements, me_data.Name
}

func GetRedditClient(user *RedditUser) *RedditClient {
	// an auth token already exists
	if user.AuthToken != "" {
		return NewAuthenticatedClient(getAppName(), user.AuthToken)
	} else {
		client, _ := NewClient(getAppName(), getAppId(), getAppSecret(), user.Username, user.Password)
		user.AuthToken = client.auth_token
		return client
	}
}

func GetUsers() []RedditUser {
	// TODO: fill this up with actual information from the data base
	return []RedditUser{
		{
			Username: getLocalUserName(),
			Password: getLocalUserPw(),
		},
	}
}
