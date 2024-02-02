package reddit

import (
	"encoding/json"
	"os"
)

type ContentItem struct {
	// unique identifier across media source. every reddit item has one. In reddit this is the name
	// in azure cosmos DB every record/item has to have an id.
	// In case of media content the media content itself comes with an unique identifier that we can use
	// GlobalId string `json:"global_id,omitempty"`

	// which social media source is this coming from
	Source string `json:"source" bson:"source"`
	// unique id across Source
	Id string `json:"cid" bson:"cid"`

	// represents text title of the item. Applies to subreddits and posts but not comments
	Title string `json:"title,omitempty" bson:"title,omitempty"`
	// unique short name across the Source
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	// Subreddit, Post or Comment. This is not directly serialized
	Kind string `json:"kind" bson:"kind"`

	// Applies to comments and posts.
	// For comments: this represents which post or comment does this comment respond to.
	// for posts: this is the same value as the channel
	ChannelName string `json:"channel,omitempty" bson:"channel,omitempty"`

	//post text
	Text string `json:"text" bson:"text"`
	// for posts this is url posted by the post
	// for subreddit this is link
	Url string `json:"url,omitempty" bson:"url,omitempty"`

	//subreddit category
	Category string `json:"category,omitempty" bson:"category,omitempty"`

	// author of posts or comments. Empty for subreddits
	Author string `json:"author,omitempty" bson:"author,omitempty"`
	// date of creation of the post or comment. Empty for subreddits
	Created float64 `json:"created,omitempty" bson:"created,omitempty"`

	// Applies to posts and comments. Doesn't apply to subreddits
	Score int `json:"score,omitempty" bson:"score,omitempty"`
	// Number of comments to a post or a comment. Doesn't apply to subreddit
	Comments int `json:"comments,omitempty" bson:"comments,omitempty"`
	// Number of subscribers to a channel (subreddit). Doesn't apply to posts or comments
	Subscribers int `json:"subscribers,omitempty" bson:"subscribers,omitempty"`
	// number of likes, claps, thumbs-up
	ThumbsupCount int `json:"likes,omitempty" bson:"likes,omitempty"`
	// Applies to subreddit posts and comments. Doesn't apply to subreddits
	ThumbsupRatio float64 `json:"likes_ratio,omitempty" bson:"likes_ratio,omitempty"`

	Digest string `json:"digest,omitempty" bson:"digest,omitempty"`
}

type EngagementItem struct {
	// in cosmos DB every item has to have an id. Here the id will be synthetic
	// other than azure cosmos DB literally no one cares about this field
	// RecordId      string `json:"id"`
	// ContentId     string `json:"content_id"`
	// Source        string `json:"source"`
	// UserId        string `json:"user_id"`
	// Processed     bool   `json:"processed,omitempty"`
	// Action        string `json:"action,omitempty"`
	// ActionContent string `json:"content,omitempty"`
	Username string `json:"username"`
	Source   string `json:"source"`
	Id       string `json:"cid"`
	Action   string `json:"action"`
}

func NewContents(contents []*ContentItem) {
	_temp_writeJsonFile("contents", contents)
}

func NewEngagements(engagements []*EngagementItem) {
	_temp_writeJsonFile("engagements", engagements)
}

func _temp_writeJsonFile(file_name string, data any) {
	json_data, _ := json.Marshal(data)
	os.WriteFile(file_name+".json", json_data, 0644)
}

func GetRedditUsers() []RedditUser {
	// TODO: fill this up with actual information from the data base
	return []RedditUser{
		{
			Username: getLocalUserName(),
			Password: getLocalUserPw(),
		},
	}
}

// internal text utility functions
