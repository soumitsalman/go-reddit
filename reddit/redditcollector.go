package reddit

import (
	"fmt"
	"strings"
)

const (
	MIN_SUBSCRIBER_LIMIT = 1000
	MAX_POST_LIMIT       = 10
	// MAX_CHILDREN_LIMIT   = 5
	MAX_DIGEST_LENGTH = 6144 * 4 // 6144 tokens are roughly 6144*4 characters. This is around 7.5 pages full of content

)

const (
	REDDIT_URL    = "http://www.reddit.com"
	REDDIT_SOURCE = "REDDIT"
)

func MapToArray[TKey comparable, TValue any](list map[TKey]TValue) ([]TKey, []TValue) {
	keys := make([]TKey, 0, len(list))
	values := make([]TValue, 0, len(list))
	for key, val := range list {
		keys = append(keys, key)
		values = append(values, val)
	}
	return keys, values
}

func AppendMaps[TKey comparable, TValue any](to_map, from_map map[TKey]TValue) map[TKey]TValue {
	for key, val := range from_map {
		to_map[key] = val
	}
	return to_map
}

func CollectAllUserItems() ([]*ContentItem, []*EngagementItem) {
	var temp_contents = make(map[string]*ContentItem)
	var engagements = make([]*EngagementItem, 0, 5)

	for _, user := range GetRedditUsers() {
		user_contents, user_engagements := collectOneUserItems_map(&user)
		temp_contents = AppendMaps(temp_contents, user_contents)
		_, temp_engagements := MapToArray[string, *EngagementItem](user_engagements)
		engagements = append(engagements, temp_engagements...)
	}

	_, contents := MapToArray[string, *ContentItem](temp_contents)
	// log.Println(len(contents), "contents collected", "|", len(engagements), "engagements collected")

	return contents, engagements
}

func CollectOneUserItems(user *RedditUser) ([]*ContentItem, []*EngagementItem) {
	temp_contents, temp_engagements := collectOneUserItems_map(user)
	_, contents := MapToArray[string, *ContentItem](temp_contents)
	_, engagements := MapToArray[string, *EngagementItem](temp_engagements)
	return contents, engagements
}

func collectOneUserItems_map(user *RedditUser) (map[string]*ContentItem, map[string]*EngagementItem) {
	// instantiate client for data collection
	client := newRedditClient(user)
	if client == nil {
		return nil, nil
	}

	var user_contents = make(map[string]*ContentItem)
	var user_engagements = make(map[string]*EngagementItem)

	collect := func(reddit_item *RedditItem, collect_similar bool) []RedditItem {
		//check cache
		if _, ok := user_contents[reddit_item.Name]; !ok {
			ds_item, eng_item, children := collectRedditItem(client, reddit_item, collect_similar)
			user_contents[reddit_item.Name] = ds_item
			if eng_item != nil {
				user_engagements[reddit_item.Name] = eng_item
			}
			return children
		}
		return nil
	}

	var subreddits, _ = client.Subreddits()
	for _, sr := range subreddits {
		children := collect(&sr, true)
		var post_remaining = MAX_POST_LIMIT
		for _, child := range children {
			// collect if its a HOt POST with top or if its POST
			// collect if its a SUBREDDIT with at least min-subscribers
			if child.Kind == POST && post_remaining > 0 {
				post_remaining -= 1
				collect(&child, false)
			} else if child.Kind == SUBREDDIT && child.NumSubscribers >= MIN_SUBSCRIBER_LIMIT {
				collect(&child, false)
			}
		}
	}

	return user_contents, user_engagements
}

func collectRedditItem(client *RedditClient, item *RedditItem, collect_similar bool) (*ContentItem, *EngagementItem, []RedditItem) {
	var content_item *ContentItem
	var children []RedditItem
	// if it is a subreddit then get the top X posts
	switch item.Kind {
	case SUBREDDIT:
		// load the hot posts in this subreddit
		posts, _ := client.Posts(item, HOT)
		// log.Println(len(posts), "HOT posts collected for", item.DisplayNamePrefixed)
		content_item = newContentItem(item, posts) // safe_slice(children, 0, MAX_CHILDREN_LIMIT))
		if collect_similar {
			// now collect the similar subreddits as well to return as part of the RedditItems to explore
			similar, _ := client.SimilarSubreddits(item)
			// log.Println(len(similar), "similar subreddits collected for", item.DisplayNamePrefixed)
			children = append(posts, similar...)
		} else {
			children = posts
		}
	default:
		// retrieve comments from this post
		comments, _ := client.RetrieveComments(item)
		// log.Println(len(comments), "comments collected for", item.Name, "in", item.SubredditPrefixed)
		content_item = newContentItem(item, comments) // safe_slice(comments, 0, MAX_CHILDREN_LIMIT))
	}

	return content_item, newEngagementItem(client.User, item), children
}

func newRedditClient(user *RedditUser) *RedditClient {
	// an auth token already exists
	if user.AuthToken != "" {
		return NewAuthenticatedClient(getAppName(), user.AuthToken)
	} else {
		client, _ := NewClient(getAppName(), getAppId(), getAppSecret(), user.Username, user.Password)
		return client
	}
}

func newContentItem(item *RedditItem, children []RedditItem) *ContentItem {
	// special case arbiration functions
	subscribers := func() int {
		switch item.Kind {
		case SUBREDDIT:
			return item.NumSubscribers
		default:
			return item.SubredditSubscribers
		}
	}

	category := func() string {
		switch item.Kind {
		case SUBREDDIT:
			return item.SubredditCategory
		default:
			return item.PostCategory
		}
	}

	channel := func() string {
		if item.Kind == SUBREDDIT {
			return item.DisplayNamePrefixed
		}
		return item.SubredditPrefixed
	}

	url := func() string {
		if item.Kind == SUBREDDIT {
			return REDDIT_URL + item.Url
		}
		return REDDIT_URL + item.Link
	}

	kind := func() string {
		if item.Kind == SUBREDDIT {
			return "channel"
		}
		return item.Kind
	}

	digest := func() string {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("%s: %s\n\n", item.Kind, item.ExtractedText))
		for _, child := range children {
			builder.WriteString(fmt.Sprintf("%s: %s\n\n", child.Kind, child.ExtractedText))
			if builder.Len() >= MAX_DIGEST_LENGTH {
				// it will overflow a bit but thats okay since embeddings does its own truncation
				break
			}
		}
		return builder.String()
	}

	// create the top level instance for item
	return &ContentItem{
		Source:        REDDIT_SOURCE,
		Id:            item.Name,
		Title:         item.Title,
		Kind:          kind(),
		Name:          item.DisplayName,
		ChannelName:   channel(),
		Text:          item.ExtractedText,
		Category:      category(),
		Url:           url(), // appending www.reddit.com
		Author:        item.Author,
		Created:       item.CreatedDate,
		Score:         item.Score,
		Comments:      item.NumComments,
		Subscribers:   subscribers(),
		ThumbsupCount: item.Ups,
		ThumbsupRatio: item.UpvoteRatio,
		Digest:        digest(),
	}
}

func newEngagementItem(user *RedditUser, item *RedditItem) *EngagementItem {
	eng_item := &EngagementItem{
		Username: user.Username,
		Source:   REDDIT_SOURCE,
		Id:       item.Name,
	}
	switch item.Kind {
	case SUBREDDIT:
		if item.UserIsContributor || item.UserIsSubscriber || item.UserIsModerator {
			eng_item.Action = "joined"
			return eng_item
		}
	default:
		if user.Username == item.Author {
			eng_item.Action = "authored"
			return eng_item
		}
	}
	return nil
}

func safeSlice[T any](array []T, start, end int) []T {
	if array == nil {
		return array
	}
	if end < 0 || end > len(array) {
		end = len(array)
	}
	if start > len(array) {
		start = len(array)
	}
	return array[start:end]
}
