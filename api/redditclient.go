package api

import (
	// "log"

	"fmt"
	"log"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const (
	REDDIT_OAUTH_AUTHORIZE_URL = "https://www.reddit.com/api/v1/authorize"
	REDDIT_OAUTH_URL           = "https://www.reddit.com/api/v1/access_token"
	REDDIT_DATA_URL            = "https://oauth.reddit.com"
)

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
	Kind          string // Subreddit, Post or Comment. This is not directly serialized
	ExtractedText string // This is the extracted text after stripping out the HTML tags and collecting contents in an URL. This is not directly serialized from Reddit but rather computed

	Name                  string  `json:"name"`         // unique identifier across media source. every reddit item has one
	DisplayName           string  `json:"display_name"` // url name for subreddits
	DisplayNamePrefixed   string  `json:"display_name_prefixed"`
	Id                    string  `json:"id"`        // unique identifier across item Kind
	Title                 string  `json:"title"`     // represents text title of the item. Applies to subreddits and posts but not comments
	Subreddit             string  `json:"subreddit"` // display_name of the subreddit where the post or comment is in
	SubredditPrefixed     string  `json:"subreddit_name_prefixed"`
	Parent                string  `json:"link_id"`                 // For comments: this represents which post or comment does this comment respond to. for posts: this is the same value as the channel
	CommentBodyHtml       string  `json:"body_html"`               // comment body
	PostTextHtml          string  `json:"selftext_html"`           // post text
	Url                   string  `json:"url"`                     // for posts this is url posted by the post. for subreddit this is clickable link
	PublicDescriptionHtml string  `json:"public_description_html"` //subreddit short description
	DescriptionHtml       string  `json:"description_html"`        //subreddit long description
	SubredditCategory     string  `json:"advertiser_category"`     //subreddit category
	PostCategory          string  `json:"link_flair_text"`         // optional author or creator defined category of the post topic or subreddit topic
	Link                  string  `json:"permalink"`               // url or link to the post or comment. For subreddits this would be the URL field
	Author                string  `json:"author"`                  // author of posts or comments. Empty for subreddits
	CreatedDate           float64 `json:"created"`                 // date of creation of the post or comment. Empty for subreddits

	Score                int     `json:"score,omitempty"`       // Applies to posts and comments. Doesn't apply to subreddits
	NumComments          int     `json:"num_comments"`          // Number of comments to a post or a comment. Doesn't apply to subreddit
	NumSubscribers       int     `json:"subscribers"`           // Number of subscribers to a channel (subreddit). Doesn't apply to posts or comments
	SubredditSubscribers int     `json:"subreddit_subscribers"` // this applies to posts and comments to indicate the same thing as above
	Ups                  int     `json:"ups"`
	UpvoteRatio          float64 `json:"upvote_ratio"` // Applies to subreddit posts and comments. Doesn't apply to subreddits

	// collecting user specific info for subreddits
	UserIsSubscriber  bool `json:"user_is_subscriber"`
	UserIsModerator   bool `json:"user_is_moderator"`
	UserIsContributor bool `json:"user_is_contributor"`
}

// the json tags are there to accommodate serialization directly from reddit api
type RedditUser struct {
	UserId       string `json:"ignore_id,omitempty"`
	Username     string `json:"name,omitempty"`
	Password     string `json:"ignore_password,omitempty"`
	AccessToken  string `json:"ignore_access_token,omitempty"`
	RefreshToken string `json:"ignore_refresh_token,omitempty"`
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
	User        *RedditUser
}

// type AuthFailureMessage string

// func (msg AuthFailureMessage) Error() string {
// 	return string(msg)
// }

type RedditAuthenticationResult struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	FailureMessage string `json:"message"`
}

func (res RedditAuthenticationResult) Error() string {
	return res.FailureMessage
}

func NewRedditClient(user *RedditUser) (*RedditClient, error) {
	if user.RefreshToken != "" {
		// log.Println("OAUTH with refresh_token")
		auth_grant := map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": user.RefreshToken,
		}
		return authenticateRedditClient(user.UserId, auth_grant)
	} else if user.Password != "" {
		// log.Println("OAUTH with password")
		auth_grant := map[string]string{
			"grant_type": "password",
			"username":   user.Username,
			"password":   user.Password,
		}
		return authenticateRedditClient(user.UserId, auth_grant)
	} else if user.AccessToken != "" {
		// log.Println("OAUTH with auth_otken")
		return NewAuthenticatedRedditClient(user), nil
	} else {
		return nil, &RedditAuthenticationResult{FailureMessage: "Insufficient Input Parameters. Needs either RefreshToken, Username+Password or existing AuthToken."}
	}
}

func NewOauthRedditClient(user_id, code string) (*RedditClient, error) {
	auth_grant := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": GetOauthRedirectUri(),
	}
	return authenticateRedditClient(user_id, auth_grant)
}

func NewAuthenticatedRedditClient(user *RedditUser) *RedditClient {
	return &RedditClient{
		User: user,
		http_client: resty.New().
			SetTimeout(MAX_WAIT_TIME).
			SetBaseURL(REDDIT_DATA_URL).
			SetHeader("User-Agent", getUserAgent()).
			SetAuthToken(user.AccessToken),
	}
}

func authenticateRedditClient(user_id string, auth_grant map[string]string) (*RedditClient, error) {
	var oauth_result RedditAuthenticationResult
	resty.New().R().
		SetBasicAuth(GetAppId(), GetAppSecret()).
		SetHeader("User-Agent", GetAppName()).
		SetHeader("Content-Type", URL_ENCODED_BODY).
		SetFormData(auth_grant).
		SetResult(&oauth_result).
		SetError(&oauth_result).
		Post(REDDIT_OAUTH_URL)

	if oauth_result.FailureMessage != "" {
		log.Println("Authentication Failed")
		return nil, &oauth_result
	}

	log.Println("Authentication Succeeded for", user_id)
	client := NewAuthenticatedRedditClient(&RedditUser{UserId: user_id, AccessToken: oauth_result.AccessToken, RefreshToken: oauth_result.RefreshToken})

	// add username
	if me_data, err := client.Me(); err == nil {
		client.User.Username = me_data.Username
	}
	return client, nil
}

func (client *RedditClient) Me() (RedditUser, error) {
	var me_data RedditUser
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
	return listing_data.getItems(SUBREDDIT), nil
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
	return listing.getItems(SUBREDDIT), nil
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
	return listing.getItems(SUBREDDIT), nil
}

// gets posts: hot, best and top depending what is specified through post_type
// if sub_reddit display name is not specified it will pull from the overall list of posts instead of a specific subreddit
// TODO: deal with paging
func (client *RedditClient) Posts(subreddit *RedditItem, post_type string) ([]RedditItem, error) {
	var url string
	// if subreddit is NOT nil, pull in the post_type posts from the subreddit
	// or else pull in post from users top profile
	if subreddit != nil {
		url = fmt.Sprintf("/%s/", subreddit.DisplayNamePrefixed)
	} else {
		url = "/"
	}
	var listing listingData
	if _, err := client.http_client.R().
		SetResult(&listing).
		Get(url + post_type); err != nil {
		log.Println("failed getting posts from", url)
		return nil, err
	}
	return listing.getItems(POST), nil
}

// retrieves comments for a specific post
func (client *RedditClient) RetrieveComments(post *RedditItem) ([]RedditItem, error) {
	// this returns multiple listings
	var listing []listingData
	if _, err := client.http_client.R().
		SetResult(&listing).
		Get(fmt.Sprintf("/%s/comments/%s", post.SubredditPrefixed, post.Id)); err != nil {
		log.Println("error pulling in comments", err)
		return nil, err
	}

	var collection []RedditItem
	for _, listing := range listing {
		collection = append(collection, listing.getItems(COMMENT)...)
	}
	return collection, nil
}

func GetRedditAuthorizationUrl(user_id string) string {
	params := url.Values{}
	params.Add("client_id", GetAppId())
	params.Add("response_type", "code")
	params.Add("state", user_id)
	params.Add("redirect_uri", GetOauthRedirectUri())
	params.Add("duration", "permanent")
	params.Add("scope", "identity edit read")

	return fmt.Sprintf("%s?%s", REDDIT_OAUTH_AUTHORIZE_URL, params.Encode())
}

// internal utility functions

func (listing_data *listingData) getItems(kind string) []RedditItem {
	items := make([]RedditItem, len(listing_data.Data.Children))
	var counter int = 0
	for _, v := range listing_data.Data.Children {
		item_kind := extractKind(v.Kind)
		// check if the item is of the kind that is expected
		if kind == "*" || kind == item_kind {
			items[counter] = v.Data
			items[counter].Kind = kind
			counter += 1
		}
	}
	return items[0:counter]
}

func extractKind(kind string) string {
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
