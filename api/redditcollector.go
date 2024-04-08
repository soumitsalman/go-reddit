package api

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	ds "github.com/soumitsalman/beansack/sdk"
	datautils "github.com/soumitsalman/data-utils"
	dl "github.com/soumitsalman/document-loader/loaders"
	oldds "github.com/soumitsalman/media-content-service/api"
)

const (
	MIN_SUBSCRIBER_LIMIT = 10000
	MAX_POST_LIMIT       = 10
)

const (
	MAX_SUBREDDIT_TEXT_LENGTH = 1024 * 4
	MAX_POST_TEXT_LENGTH      = 3072 * 4
	MAX_ARTICLE_TEXT_LENGTH   = 4096 * 4
	MAX_COMMENT_TEXT_LENGTH   = 512 * 4

	MAX_EXTRACTED_TEXT_LENGTH = 4096 * 4
	MAX_CHILD_TEXT_LENGTH     = 512 * 4
	MIN_TEXT_LENGTH           = 5 * 4 // anything below this text length, just ignore it

	MAX_DIGEST_TEXT_LENGTH = 6144 * 4 // 6144 tokens are roughly 6144*4 characters. This is around 7.5 pages full of content
)

const (
	REDDIT_URL    = "http://www.reddit.com"
	REDDIT_SOURCE = "REDDIT"
)

const (
	_MASTER_COLLECTOR = "__DEFAULT_MASTER_COLLECTOR__"
	_UNKNOWN          = "__UNKNOWN__"
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

// COLLECTION RELATED FUNCTIONS
func CollectAndStoreAll() {
	for _, user := range GetRedditUsers() {
		user.CollectAndStore()
		// time.Sleep(MAX_WAIT_TIME) // wait out for a bit to avoid rate limiting
	}
}

func (user *RedditUser) CollectAndStore() {
	beans, noises, engagements := user.Collect()
	if len(beans) > 0 {
		StoreBeans(beans)
		StoreMediaNoises(noises)
		if user.UserId != _MASTER_COLLECTOR {
			StoreNewEngagements(engagements)
		}
		log.Printf("Finished storing for u/%s\n", user.Username)
	}
}

func (user *RedditUser) Collect() ([]*ds.Bean, []*ds.BeanMediaNoise, []*oldds.UserEngagementItem) {
	client, err := NewRedditClient(user)
	if err != nil {
		return nil, nil, nil
	}

	var beans, media_noises, engagements = make(map[string]*ds.Bean), make(map[string]*ds.BeanMediaNoise), make(map[string]*oldds.UserEngagementItem)
	collect := func(reddit_item *RedditItem, collect_similar bool) []RedditItem {
		//check cache
		if _, ok := beans[reddit_item.Name]; !ok {
			bean, media_noise, eng, children := collectRedditItem(client, reddit_item, collect_similar)
			// if we can't build a digest then we will not send it
			if len(bean.Text) >= MIN_TEXT_LENGTH {
				beans[reddit_item.Name] = bean
				media_noises[reddit_item.Name] = media_noise
			}
			if eng != nil {
				engagements[reddit_item.Name] = eng
			}
			return children
		}
		return nil
	}

	log.Printf("Starting collection for u/%s\n", client.User.Username)

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

	_, res_beans := datautils.MapToArray[string, *ds.Bean](beans)
	_, res_media_noises := datautils.MapToArray(media_noises)
	_, res_engagements := datautils.MapToArray[string, *oldds.UserEngagementItem](engagements)

	log.Printf("Finished collection for u/%s | %d contents, %d engagements\n", client.User.Username, len(res_beans), len(res_engagements))
	return res_beans, res_media_noises, res_engagements
}

func collectRedditItem(client *RedditClient, item *RedditItem, collect_similar bool) (*ds.Bean, *ds.BeanMediaNoise, *oldds.UserEngagementItem, []RedditItem) {
	var bean = item.toBean(nil)
	var media_noise *ds.BeanMediaNoise
	var children []RedditItem
	// if it is a subreddit then get the top X posts
	switch item.Kind {
	case SUBREDDIT:
		// load the hot posts in this subreddit
		posts, _ := client.Posts(item, HOT)
		// log.Println(len(posts), "HOT posts collected for", item.DisplayNamePrefixed)
		media_noise = item.toBeanMediaNoise(posts)

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
		media_noise = item.toBeanMediaNoise(comments) // safe_slice(comments, 0, MAX_CHILDREN_LIMIT))
	}

	return bean, media_noise, item.toUserEngagement(client.User), children
}

// DATA FORMAT TRANSFORMERS
func (item *RedditItem) toBean(children []RedditItem) *ds.Bean {
	// create the top level instance for item
	return &ds.Bean{
		Url:      item.contentUrl(),
		Source:   REDDIT_SOURCE,
		Title:    item.Title,
		Kind:     item.kind(),
		Text:     item.extractedText(),
		Author:   item.Author,
		Created:  int64(item.CreatedDate),
		Keywords: item.category(),
	}
}

func (item *RedditItem) toBeanMediaNoise(children []RedditItem) *ds.BeanMediaNoise {
	// special case arbiration functions
	subscribers := func() int {
		switch item.Kind {
		case SUBREDDIT:
			return item.NumSubscribers
		default:
			return item.SubredditSubscribers
		}
	}

	channel := func() string {
		if item.Kind == SUBREDDIT {
			return item.DisplayNamePrefixed
		}
		return item.SubredditPrefixed
	}

	// create the top level instance for item
	return &ds.BeanMediaNoise{
		BeanUrl:       item.contentUrl(),
		Media:         REDDIT_SOURCE,
		ContentId:     item.Name,
		Name:          item.DisplayNamePrefixed,
		MediaChannel:  channel(),
		MediaUrl:      item.containerUrl(),
		Author:        item.Author,
		Score:         item.Score,
		Comments:      item.NumComments,
		Subscribers:   subscribers(),
		ThumbsupCount: item.Ups,
		ThumbsupRatio: item.UpvoteRatio,
		// Digest:        item.childrenDigest(children),
	}
}

func (item *RedditItem) toUserEngagement(user *RedditUser) *oldds.UserEngagementItem {
	eng_item := &oldds.UserEngagementItem{
		Username:   user.Username,
		UserSource: REDDIT_SOURCE,
		Source:     REDDIT_SOURCE,
		ContentId:  item.Name,
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

// FIELD EXTRACTION FUNCTIONS
func (item *RedditItem) category() []string {
	var res string
	switch item.Kind {
	case SUBREDDIT:
		res = strings.TrimSpace(item.SubredditCategory)
	default:
		res = strings.TrimSpace(item.PostCategory)
	}

	if len(res) > 0 {
		return []string{res}
	} else {
		return nil
	}
}

func (item *RedditItem) contentUrl() string {
	switch item.kind() {
	case ds.ARTICLE:
		return item.Url
	case ds.CHANNEL:
		return REDDIT_URL + item.Url
	default:
		return REDDIT_URL + item.Link
	}
}

func (item *RedditItem) containerUrl() string {
	switch item.kind() {
	case ds.ARTICLE:
		return REDDIT_URL + item.Link
	case ds.CHANNEL:
		return REDDIT_URL + item.Url
	default:
		return REDDIT_URL + item.Link
	}
}

func (item *RedditItem) kind() string {
	switch item.Kind {
	case SUBREDDIT:
		return ds.CHANNEL
	case POST:
		if item.PostTextHtml != "" {
			return ds.POST
		} else if item.Url != "" {
			return ds.ARTICLE
		} else {
			return _UNKNOWN
		}
	case COMMENT:
		return ds.COMMENT
	default:
		return _UNKNOWN
	}
}

func (item *RedditItem) digest(children []RedditItem) string {
	var builder strings.Builder
	var body_text string

	switch item.kind() {
	// for subreddits the description doesnt matter as much as the top posts
	case SUBREDDIT:
		body_text = fmt.Sprintf("%s: %s\n\nPOSTS in this subreddit:\n", item.Kind, item.DisplayNamePrefixed)
	// if it is a post or a comment, add a part of the body
	case POST:
		body_text = fmt.Sprintf("%s: %s\n\nCOMMENTS to this post:\n", item.Kind, datautils.TruncateTextWithEllipsis(item.extractedText(), MAX_POST_TEXT_LENGTH))
	case COMMENT:
		body_text = fmt.Sprintf("%s: %s\n\nCOMMENTS to this comment:\n", item.Kind, datautils.TruncateTextWithEllipsis(item.extractedText(), MAX_COMMENT_TEXT_LENGTH))
	}

	builder.WriteString(body_text)
	for _, child := range children {
		child_text := datautils.TruncateTextWithEllipsis(child.extractedText(), MAX_CHILD_TEXT_LENGTH)
		if len(child_text) >= MIN_TEXT_LENGTH {
			builder.WriteString(fmt.Sprintf("%s: %s\n\n", child.Kind, child_text))
		}
		if builder.Len() >= MAX_DIGEST_TEXT_LENGTH {
			// it will overflow a bit but thats okay since embeddings does its own truncation
			break
		}
	}
	return builder.String()
}

var url_collector = dl.NewRedditLinkLoader()

func (item *RedditItem) extractedText() string {
	if item.ExtractedText == "" {
		var temp_text string
		switch item.Kind {
		case SUBREDDIT:
			temp_text = extractTextFromHtml(item.PublicDescriptionHtml + "\n" + item.DescriptionHtml)
		case POST:
			if item.PostTextHtml != "" {
				// this is a post with contents written in reddit
				temp_text = extractTextFromHtml(item.PostTextHtml)
			} else if item.Url != "" {
				// this is link to a new article posted in reddit
				temp_text = url_collector.LoadDocument(item.Url).Text
			}
		case COMMENT:
			temp_text = extractTextFromHtml(item.CommentBodyHtml)
		}
		item.ExtractedText = cleanupText(temp_text, MAX_EXTRACTED_TEXT_LENGTH)
	}

	return item.ExtractedText
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
	return datautils.TruncateTextWithEllipsis(text, max_length)
}
