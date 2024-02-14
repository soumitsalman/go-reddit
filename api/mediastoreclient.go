package api

import (
	"log"

	"github.com/go-resty/resty/v2"
	ds "github.com/soumitsalman/media-content-service/api"
)

func StoreNewContents(contents []*ds.MediaContentItem) {
	_, err := getMediaStoreClient().R().
		SetHeader("Content-Type", JSON_BODY).
		SetBody(contents).
		Post("/contents")
	if err != nil {
		log.Println("FAILED storing new contents", err)
	}
}

func StoreNewEngagements(engagements []*ds.UserEngagementItem) {
	_, err := getMediaStoreClient().R().
		SetHeader("Content-Type", JSON_BODY).
		SetBody(engagements).
		Post("/engagements")
	if err != nil {
		log.Println("FAILED storing new engagements", err)
	}
}

func GetRedditUsers() []RedditUser {
	var creds []RedditUser
	_, err := getMediaStoreClient().R().
		SetResult(&creds).
		Get("/users/REDDIT")
	if err != nil {
		log.Println("FAILED getting user creds", err)
	}

	return creds
}

var media_store_client *resty.Client

func getMediaStoreClient() *resty.Client {
	if media_store_client == nil {
		media_store_client = resty.New().
			SetTimeout(MAX_WAIT_TIME).
			SetBaseURL(getMediaStoreUrl()).
			SetHeader("User-Agent", getUserAgent()).
			SetHeader("X-API-Key", getInternalAuthToken())
	}
	return media_store_client
}

// func _temp_writeJsonFile(file_name string, data any) {
// 	json_data, _ := json.Marshal(data)
// 	os.WriteFile(file_name+".json", json_data, 0644)
// }