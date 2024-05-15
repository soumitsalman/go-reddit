package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	ds "github.com/soumitsalman/beansack/sdk"
	oldds "github.com/soumitsalman/media-content-service/api"
)

type BeansackClient struct {
	config BeansackConfig
	client *resty.Client
}

func NewBeansackClient(beansack_config BeansackConfig) *BeansackClient {
	return &BeansackClient{
		config: beansack_config,
		client: resty.New().
			SetTimeout(MAX_WAIT_TIME).
			SetBaseURL(beansack_config.BeanSackUrl).
			SetHeader("User-Agent", beansack_config.UserAgent).
			SetHeader("Content-Type", JSON_BODY).
			SetHeader("X-API-Key", beansack_config.BeanSackAPIKey),
	}
}

func (client *BeansackClient) StoreBeans(contents []*ds.Bean) {
	debug_writeJsonFile(contents)
	retry.Do(
		func() error {
			_, err := client.client.R().
				SetBody(contents).
				Put("/beans")

			if err != nil {
				log.Println("FAILED storing new contents", err)
			}
			return err
		},
		retry.Attempts(3),
		retry.Delay(time.Second),
	)
}

func (client *BeansackClient) StoreNewEngagements(engagements []*oldds.UserEngagementItem) {
	// debug_writeJsonFile(engagements)
	// _, err := getMediaStoreClient().R().
	// 	SetHeader("Content-Type", JSON_BODY).
	// 	SetBody(engagements).
	// 	Post("/engagements")
	// if err != nil {
	// 	log.Println("FAILED storing new engagements", err)
	// }
}

var debug_filename_counter = 0

func debug_writeJsonFile(data any) {
	debug_filename_counter += 1
	file_name := fmt.Sprintf("%T_%d.json", data, debug_filename_counter)
	json_data, _ := json.Marshal(data)
	os.WriteFile(file_name, json_data, 0644)
}
