package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	ds "github.com/soumitsalman/beansack/sdk"
	oldds "github.com/soumitsalman/media-content-service/api"
)

func StoreBeans(contents []*ds.Bean) {
	debug_writeJsonFile(contents)
	_, err := getMediaStoreClient().R().
		SetHeader("Content-Type", JSON_BODY).
		SetBody(contents).
		Put("/beans")
	if err != nil {
		log.Println("FAILED storing new contents", err)
	}
}

func StoreNewEngagements(engagements []*oldds.UserEngagementItem) {
	// debug_writeJsonFile(engagements)
	// _, err := getMediaStoreClient().R().
	// 	SetHeader("Content-Type", JSON_BODY).
	// 	SetBody(engagements).
	// 	Post("/engagements")
	// if err != nil {
	// 	log.Println("FAILED storing new engagements", err)
	// }
}

var media_store_client *resty.Client

func getMediaStoreClient() *resty.Client {
	if media_store_client == nil {
		media_store_client = resty.New().
			SetTimeout(MAX_WAIT_TIME).
			SetBaseURL(getBeanUrl()).
			SetHeader("User-Agent", getUserAgent())
	}
	return media_store_client
}

var debug_filename_counter = 0

func debug_writeJsonFile(data any) {
	debug_filename_counter += 1
	file_name := fmt.Sprintf("%T_%d.json", data, debug_filename_counter)
	json_data, _ := json.Marshal(data)
	os.WriteFile(file_name, json_data, 0644)
}
