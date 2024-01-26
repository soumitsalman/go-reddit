package main

import (
	// "log"
	"fmt"
	"log"
	"net/url"
	"runtime"

	"github.com/go-resty/resty/v2"
)

const REDDIT_DATA_URL = "https://oauth.reddit.com"
const REDDIT_OAUTH_URL = "https://www.reddit.com/api/v1/access_token"

// internal wrapper data structure to ease json marshalling and unmarshalling
type listingData struct {
	Data struct {
		Children []struct {
			Kind string     `json:"kind"`
			Data RedditItem `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// represents Subreddit, Posts, Comments
type RedditItem struct {
	Name        string `json:"name"`         // unique identifier across media source. every reddit item has one
	DisplayName string `json:"display_name"` // url name for subreddits
	Id          string `json:"id"`           // unique identifier across item Kind
	Title       string `json:"title"`        // represents text title of the item. Applies to subreddits and posts but not comments
	// Subreddit, Post or Comment. This is not directly serialized
	Kind string

	// display_name of the subreddit where the post or comment is in
	Subreddit string `json:"subreddit"`
	// Applies to comments and posts.
	// For comments: this represents which post or comment does this comment respond to.
	// for posts: this is the same value as the channel
	Parent string `json:"link_id"`

	// comment body
	CommentBody string `json:"body_html"`
	// post text
	PostText string `json:"selftext_html"`
	// for posts this is url posted by the post
	// for subreddit this is link
	Url string `json:"url"`
	//subreddit short description
	SubredditPublicDescription string `json:"public_description_html"`
	//subreddit long description
	SubredditDescription string `json:"description_html"`
	//subreddit category
	SubredditCategory string `json:"advertiser_category"`
	// optional author or creator defined category of the post topic or subreddit topic
	PostCategory string `json:"link_flair_text"`
	// url or link to the content item in the media source
	Link string `json:"permalink"`

	// author of posts or comments. Empty for subreddits
	Author string `json:"author"`
	// date of creation of the post or comment. Empty for subreddits
	CreatedDate float64 `json:"created"`

	// Applies to posts and comments. Doesn't apply to subreddits
	Score int `json:"score"`
	// Number of comments to a post or a comment. Doesn't apply to subreddit
	NumComments int `json:"num_comments"`
	// Number of subscribers to a channel (subreddit). Doesn't apply to posts or comments
	NumSubscribers int `json:"subscribers"`
	// this applies to posts and comments to indicate the same thing as above
	SubredditSubscribers int `json:"subreddit_subscribers"`
	// Applies to subreddit posts and comments. Doesn't apply to subreddits
	UpvoteRatio float64 `json:"upvote_ratio"`

	// collecting user specific info
	UserIsSubscriber  bool `json:"user_is_subscriber"`
	UserIsModerator   bool `json:"user_is_moderator"`
	UserIsContributor bool `json:"user_is_contributor"`
}

const (
	SUBREDDIT = "subreddit"
	POST      = "post"
	COMMENT   = "comment"
)

const (
	HOT  = "hot"
	TOP  = "top"
	BEST = "best"
)

type RedditClient struct {
	http_client *resty.Client
	auth_token  string
}

type AuthFailureMessage string

func (msg AuthFailureMessage) Error() string {
	return string(msg)
}

// authenticate the user and returns a retrieval client
func NewClient(app_name, app_id, app_secret, user_name, user_pw string) (*RedditClient, error) {
	unpw := url.Values{}
	unpw.Set("grant_type", "password")
	unpw.Set("username", user_name)
	unpw.Set("password", user_pw)

	var auth_result struct {
		AccessToken    string `json:"access_token"`
		FailureMessage string `json:"message"`
	}

	resty.New().R().
		SetBasicAuth(app_id, app_secret).
		SetHeader("User-Agent", getUserAgent(app_name)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(unpw.Encode()).
		SetResult(&auth_result).
		SetError(&auth_result).
		Post(REDDIT_OAUTH_URL)

	if auth_result.FailureMessage != "" {
		log.Println("Authentication Failed")
		return nil, AuthFailureMessage(auth_result.FailureMessage)
	} else {
		log.Println("Authentication Succeeded")
		return NewAuthenticatedClient(app_name, auth_result.AccessToken), nil
	}
}

// create a client from an existing auth token
func NewAuthenticatedClient(app_name, auth_token string) *RedditClient {
	return &RedditClient{
		auth_token: auth_token,
		http_client: resty.New().
			SetBaseURL(REDDIT_DATA_URL).
			SetHeader("User-Agent", getUserAgent(app_name)).
			SetAuthToken(auth_token),
	}
}

func (client *RedditClient) Me() (RedditItem, error) {
	var me_data RedditItem
	if _, err := client.http_client.R().
		SetResult(&me_data).
		Get("/api/v1/me"); err != nil {
		return me_data, err
	}
	return me_data, nil
}

// gets subreddits that the user in the client has already subscribed to
func (client *RedditClient) Subreddits() ([]RedditItem, error) {
	var listing_data listingData
	if _, err := client.http_client.R().
		SetResult(&listing_data).
		Get("/subreddits/mine/subscriber"); err != nil {
		return nil, err
	}
	return listing_data.getItems(), nil
}

// get subreddits based on a given
// does not return unique list of items and may have duplicates
// TODO: deal with paging
func (client *RedditClient) SimilarSubreddits(subreddit *RedditItem) ([]RedditItem, error) {
	var listing listingData
	if _, err := client.http_client.R().
		SetQueryParam("sr_fullnames", subreddit.Name).
		SetResult(&listing).
		Get("/api/similar_subreddits"); err != nil {
		return nil, err
	}
	return listing.getItems(), nil
}

// uses the query string to look for sub-reddits
// min_users is used to filter for sub-reddits that has at least min_users number of users
// TODO: deal with paging
func (client *RedditClient) SubredditSearch(search_query string) ([]RedditItem, error) {
	var listing listingData
	if _, err := client.http_client.R().
		SetQueryParam("q", search_query).
		SetResult(&listing).
		Get("/subreddits/search"); err != nil {
		return nil, err
	}
	return listing.getItems(), nil
}

// gets posts: hot, best and top depending what is specified through post_type
// if sub_reddit display name is not specified it will pull from the overall list of posts instead of a specific subreddit
// TODO: deal with paging
func (client *RedditClient) Posts(subreddit *RedditItem, post_type string) ([]RedditItem, error) {
	var url string = "/"
	// pull in the post_type posts from the user's profile
	if subreddit == nil {
		url = fmt.Sprintf("r/%s/", subreddit.DisplayName)
	}
	var listing listingData
	if _, err := client.http_client.R().
		SetResult(&listing).
		Get(url + post_type); err != nil {
		return nil, err
	}
	return listing.getItems(), nil
}

// retrieves comments for a specific post
func (client *RedditClient) RetrieveComments(post *RedditItem) ([]RedditItem, error) {
	// this returns multiple listings
	var listing []listingData
	if _, err := client.http_client.R().
		SetResult(&listing).
		Get(fmt.Sprintf("/r/%s/comments/%s", post.Subreddit, post.Id)); err != nil {
		return nil, err
	}

	var collection []RedditItem
	for _, listing := range listing {
		collection = append(collection, listing.getItems()...)
	}
	return collection, nil
}

// internal utility functions

func getUserAgent(app_name string) string {
	//Windows:My Reddit Bot:1.0 (by u/botdeveloper)
	return fmt.Sprintf("%v:%v:v0.1 (by u/randomizer_000)", runtime.GOOS, app_name)
}

func (listing_data *listingData) getItems() []RedditItem {
	items := make([]RedditItem, len(listing_data.Data.Children))
	for i, v := range listing_data.Data.Children {
		items[i] = v.Data
		items[i].Kind = translateItemKind(v.Kind)
	}
	return items
}

func translateItemKind(kind string) string {
	switch kind {
	case "t5":
		return SUBREDDIT
	case "t3":
		return POST
	case "t1":
		return COMMENT
	default:
		return kind
	}
}

// type authResult struct {
// 	AccessToken    string `json:"access_token"`
// 	FailureMessage string `json:"message"`
// }

// func (res *authResult) Error() string {
// 	return res.FailureMessage
// }
