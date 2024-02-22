package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-shiori/go-readability"
	utils "github.com/soumitsalman/data-utils"
)

const (
	_MAX_CACHE_LENGTH = 50000
	_USER_AGENT       = "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.10136"
)

type UrlCollector struct {
	cache   map[string]string
	ignores []string
}

func NewUrlCollector(ignore []string) *UrlCollector {
	return &UrlCollector{
		cache:   make(map[string]string),
		ignores: ignore,
	}
}

// extracts texts from url
func (collector *UrlCollector) GetText(url string) string {
	// return from the cache
	text, ok := collector.cache[url]
	if ok {
		return text
	}

	if utils.Any[string](collector.ignores, func(skip_url *string) bool {
		return strings.HasPrefix(url, *skip_url) || strings.HasSuffix(url, *skip_url) || strings.Contains(url, *skip_url)
	}) {
		// no need to save the ones to ignore. This does not save much of compute time but frees up cache
		return ""
	}

	// this being done to skip bot detection
	client := &http.Client{Timeout: MAX_WAIT_TIME}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", _USER_AGENT)
	req.Header.Set("Accept", "text/html")
	// then check content-type to not parse through MIME content
	if resp, err := client.Do(req); (err == nil) && (resp.StatusCode == http.StatusOK) && (strings.Contains(resp.Header.Get("Content-Type"), "text/html")) {
		// log.Println("parsing url content", url)
		article, _ := readability.FromReader(resp.Body, resp.Request.URL)
		text = article.TextContent
	} else {
		// TODO: disable the error messages
		log.Println("couldn't parse url:", url, "| err:", err)
		if resp != nil {
			log.Println("StatusCode:", resp.StatusCode, "| Content-Type:", resp.Header.Get("Content-Type"))
		}
		text = ""
	}

	collector.addToCache(url, text)
	return text
}

func (collector *UrlCollector) addToCache(url, text string) {
	// delete a part of cache if goes beyond cache size
	if len(collector.cache) >= _MAX_CACHE_LENGTH {
		delete_counter := 0
		for i := range collector.cache {
			delete(collector.cache, i)

			delete_counter += 1
			if delete_counter >= _MAX_CACHE_LENGTH>>3 {
				break
			}
		}
	}

	collector.cache[url] = text
}
