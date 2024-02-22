package api

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	datautils "github.com/soumitsalman/data-utils"
	ds "github.com/soumitsalman/media-content-service/api"
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

// func CollectItems(user *RedditUser) ([]*ds.MediaContentItem, []*ds.UserEngagementItem) {
// func (client *RedditClient) CollectItems() {
// 	temp_contents, temp_engagements := client.collectItems_map()
// 	_, contents := MapToArray[string, *ds.MediaContentItem](temp_contents)
// 	_, engagements := MapToArray[string, *ds.UserEngagementItem](temp_engagements)

// 	StoreNewContents(contents)
// 	StoreNewEngagements(engagements)
// 	// return contents, engagements
// }

func (client *RedditClient) CollectItems() ([]*ds.MediaContentItem, []*ds.UserEngagementItem) {
	if client == nil {
		return nil, nil
	}

	var user_contents, user_engagements = make(map[string]*ds.MediaContentItem), make(map[string]*ds.UserEngagementItem)
	collect := func(reddit_item *RedditItem, collect_similar bool) []RedditItem {
		//check cache
		if _, ok := user_contents[reddit_item.Name]; !ok {
			ds_item, eng_item, children := collectRedditItem(client, reddit_item, collect_similar)
			// if we can't build a digest then we will not send it
			if len(ds_item.Digest) >= MIN_TEXT_LENGTH {
				user_contents[reddit_item.Name] = ds_item
			}
			if eng_item != nil {
				user_engagements[reddit_item.Name] = eng_item
			}
			return children
		}
		return nil
	}

	log.Println("Starting collection for u/", client.User.Username)

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

	_, contents := datautils.MapToArray[string, *ds.MediaContentItem](user_contents)
	_, engagements := datautils.MapToArray[string, *ds.UserEngagementItem](user_engagements)

	log.Printf("Finished collection for u/%s | %d contents, %d engagements\n", client.User.Username, len(contents), len(engagements))

	StoreNewContents(contents)
	StoreNewEngagements(engagements)

	log.Println("Finished storing for u/", client.User.Username)

	return contents, engagements
}

func collectRedditItem(client *RedditClient, item *RedditItem, collect_similar bool) (*ds.MediaContentItem, *ds.UserEngagementItem, []RedditItem) {
	var content_item *ds.MediaContentItem
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

func NewCollectorClient(user *RedditUser) *RedditClient {
	// an auth token already exists
	if user.AuthToken != "" {
		return NewAuthenticatedRedditClient(getUserAgent(), user.AuthToken)
	} else {
		client, _ := NewRedditClient(getUserAgent(), getAppId(), getAppSecret(), user.Username, user.Password)
		return client
	}
}

func newContentItem(item *RedditItem, children []RedditItem) *ds.MediaContentItem {
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
		var body_text string

		switch item.Kind {
		// for subreddits the description doesnt matter as much as the top posts
		case SUBREDDIT:
			body_text = fmt.Sprintf("%s: %s\n\nPOSTS in this subreddit:\n", item.Kind, item.DisplayNamePrefixed)
		// if it is a post or a comment, add a part of the body
		case POST:
			body_text = fmt.Sprintf("%s: %s\n\nCOMMENTS to this post:\n", item.Kind, datautils.TruncateTextWithEllipsis(item.extractText(), MAX_POST_TEXT_LENGTH))
		case COMMENT:
			body_text = fmt.Sprintf("%s: %s\n\nCOMMENTS to this comment:\n", item.Kind, datautils.TruncateTextWithEllipsis(item.extractText(), MAX_COMMENT_TEXT_LENGTH))
		}

		builder.WriteString(body_text)
		for _, child := range children {
			child_text := datautils.TruncateTextWithEllipsis(child.extractText(), MAX_CHILD_TEXT_LENGTH)
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

	// create the top level instance for item
	return &ds.MediaContentItem{
		Source:        REDDIT_SOURCE,
		Id:            item.Name,
		Title:         item.Title,
		Kind:          kind(),
		Name:          item.DisplayName,
		ChannelName:   channel(),
		Text:          item.extractText(),
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

func newEngagementItem(user *RedditUser, item *RedditItem) *ds.UserEngagementItem {
	eng_item := &ds.UserEngagementItem{
		Username:  user.Username,
		Source:    REDDIT_SOURCE,
		ContentId: item.Name,
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

var url_collector = NewUrlCollector([]string{
	"https://v.redd.it", "https://i.redd.it", "https://www.reddit.com/gallery",
	"https://www.youtube.com",
	".png", ".jpeg", ".jpg", ".gif", ".webp",
	".mp4", ".avi", ".mkv", ".mp3", ".wav",
	".pdf",
})

func (item *RedditItem) extractText() string {
	if item.ExtractedText == "" {
		var temp_text string
		switch item.Kind {
		case SUBREDDIT:
			// item.ExtractedText = cleanupText(
			// 	extractTextFromHtml(item.PublicDescriptionHtml+"\n"+item.DescriptionHtml),
			// 	MAX_SUBREDDIT_TEXT_LENGTH)
			temp_text = extractTextFromHtml(item.PublicDescriptionHtml + "\n" + item.DescriptionHtml)

		case POST:
			if item.PostTextHtml != "" {
				// this is a post with contents written in reddit
				// item.ExtractedText = cleanupText(
				// 	extractTextFromHtml(item.PostTextHtml),
				// 	MAX_POST_TEXT_LENGTH)
				temp_text = extractTextFromHtml(item.PostTextHtml)
			} else if item.Url != "" {
				// this is link to a new article posted in reddit
				// item.ExtractedText = cleanupText(
				// 	url_collector.GetText(item.Url),
				// 	MAX_ARTICLE_TEXT_LENGTH)
				temp_text = url_collector.GetText(item.Url)
			}
		case COMMENT:
			// item.ExtractedText = cleanupText(
			// 	extractTextFromHtml(item.CommentBodyHtml),
			// 	MAX_COMMENT_TEXT_LENGTH)
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
