package api

import (
	// "log"

	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/go-shiori/go-readability"

	utils "github.com/soumitsalman/data-utils"
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
	Kind          string // Subreddit, Post or Comment. This is not directly serialized
	ExtractedText string // this applies if PostTextHtml is empty. This contains the text extracted from URL. This is not directly serialized

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

type RedditUser struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	AuthToken string `json:"auth_token"`
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

const (
	MAX_WAIT_TIME = 60 * time.Second
)

const (
	MAX_SUBREDDIT_TEXT_LENGTH = 1024 * 4
	MAX_POST_TEXT_LENGTH      = 3072 * 4
	MAX_ARTICLE_TEXT_LENGTH   = 4096 * 4
	MAX_COMMENT_TEXT_LENGTH   = 512 * 4
)

type RedditClient struct {
	http_client *resty.Client
	User        *RedditUser
}

type AuthFailureMessage string

func (msg AuthFailureMessage) Error() string {
	return string(msg)
}

// authenticate the user and returns a retrieval client
func NewRedditClient(app_id, app_secret, user_name, user_pw string) (*RedditClient, error) {
	// unpw := url.Values{}
	// unpw.Set("grant_type", "password")
	// unpw.Set("username", user_name)
	// unpw.Set("password", user_pw)
	unpw := map[string]string{
		"grant_type": "password",
		"username":   user_name,
		"password":   user_pw,
	}

	var auth_result struct {
		AccessToken    string `json:"access_token"`
		FailureMessage string `json:"message"`
	}

	resty.New().R().
		SetBasicAuth(app_id, app_secret).
		SetHeader("User-Agent", getUserAgent()).
		SetHeader("Content-Type", URL_ENCODED_BODY).
		SetFormData(unpw).
		// SetFormDataFromValues(unpw).
		// SetBody(unpw.Encode()).
		SetResult(&auth_result).
		SetError(&auth_result).
		Post(REDDIT_OAUTH_URL)

	if auth_result.FailureMessage != "" {
		log.Println("Authentication Failed")
		return nil, AuthFailureMessage(auth_result.FailureMessage)
	}

	log.Println("Authentication Succeeded for", user_name)
	client := NewAuthenticatedRedditClient(auth_result.AccessToken)
	client.User.Username = user_name
	return client, nil
}

// create a client from an existing auth token
func NewAuthenticatedRedditClient(auth_token string) *RedditClient {
	return &RedditClient{
		User: &RedditUser{AuthToken: auth_token},
		http_client: resty.New().
			SetTimeout(MAX_WAIT_TIME).
			SetBaseURL(REDDIT_DATA_URL).
			SetHeader("User-Agent", getUserAgent()).
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

// internal utility functions

func (listing_data *listingData) getItems(kind string) []RedditItem {
	items := make([]RedditItem, len(listing_data.Data.Children))
	var counter int = 0
	for _, v := range listing_data.Data.Children {
		item_kind := extractItemKind(v.Kind)
		// check if the item is of the kind that is expected
		if kind == "*" || kind == item_kind {
			items[counter] = v.Data
			items[counter].Kind = kind
			items[counter].ExtractedText = extractText(&items[counter])

			counter += 1
		}
	}
	return items[0:counter]
}

func extractText(item *RedditItem) string {
	var result string
	switch item.Kind {
	case SUBREDDIT:
		result = cleanupText(
			extractTextFromHtml(item.PublicDescriptionHtml+"\n"+item.DescriptionHtml),
			MAX_SUBREDDIT_TEXT_LENGTH)

	case POST:
		if item.PostTextHtml != "" {
			// this is a post with contents written in reddit
			result = cleanupText(
				extractTextFromHtml(item.PostTextHtml),
				MAX_POST_TEXT_LENGTH)
		} else if item.Url != "" {
			// this is link to a new article posted in reddit
			result = cleanupText(
				extractTextFromUrl(item.Url),
				MAX_ARTICLE_TEXT_LENGTH)
		}
	case COMMENT:
		result = cleanupText(
			extractTextFromHtml(item.CommentBodyHtml),
			MAX_COMMENT_TEXT_LENGTH)
	}
	return result
}

// extracts texts from url
func extractTextFromUrl(url string) string {
	// pre-emptively check urls to find the ones NOT to collect
	skip_urls := []string{
		"https://v.redd.it", "https://i.redd.it", "https://www.reddit.com/gallery",
		"https://www.youtube.com",
		".png", ".jpeg", ".jpg", ".gif", ".webp",
		".mp4", ".avi", ".mkv",
	}
	compare := func(url, skip *string) bool {
		return strings.HasPrefix(*url, *skip) || strings.HasSuffix(*url, *skip)
	}
	if utils.In[string](url, skip_urls, compare) {
		return ""
	}

	// this being done to skip bot detection
	client := &http.Client{Timeout: MAX_WAIT_TIME}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", getUserAgent())
	req.Header.Set("Accept", "text/html")
	// then check content-type to not parse through MIME content
	if resp, err := client.Do(req); (err == nil) && (resp.StatusCode == http.StatusOK) && (strings.Contains(resp.Header.Get("Content-Type"), "text/html")) {
		// log.Println("parsing url content", url)
		article, _ := readability.FromReader(resp.Body, resp.Request.URL)
		return article.TextContent
	} else {
		// TODO: disable the error messages
		log.Println("couldn't parse url:", url, "| err:", err)
		if resp != nil {
			log.Println("StatusCode:", resp.StatusCode, "| Content-Type:", resp.Header.Get("Content-Type"))
		}
		return ""
	}
}

// extract text from HTML fields
func extractTextFromHtml(content string) string {
	//there needs to be multiple runs on the NewDocumentFromReader when '<' and '>' are represented as "&lt;' and '&gt;'
	for count := 2; count > 0; count-- {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(content))
		content = doc.Text()
	}
	return content
}

func cleanupText(text string, max_length int) string {
	match_and_replace := func(text, regex_pattern, replacement string) string {
		return regexp.MustCompile(regex_pattern).ReplaceAllString(text, replacement)
	}
	// replace 2+ ' ' with 1 ' '
	// text = match_and_replace(text, "\t+", "\t") // regexp.MustCompile(`\t+`).ReplaceAllString(text, "\t")
	// replace 2+ \t with 1 \t
	// text = match_and_replace(text, " +", " ") // regexp.MustCompile(` +`).ReplaceAllString(text, " ")
	// 1 or more spaces, \n, 1 or more spaces
	text = match_and_replace(text, `\s+\n|\n\s+|\s+\n\s+`, "\n") // regexp.MustCompile(`[ ]+\n+`).ReplaceAllString(text, "\n")
	// replace 3+ \n with \n\n
	text = match_and_replace(text, "(\r?\n){3,}", "\n\n") // regexp.MustCompile(`(\r?\n){3,}`).ReplaceAllString(text, "\n\n")
	// now trim the leading and trailing spaces
	text = strings.TrimSpace(text)
	if max_length < 0 || max_length > len(text) {
		max_length = len(text)
	}
	return text[:max_length]
}

func extractItemKind(kind string) string {
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
