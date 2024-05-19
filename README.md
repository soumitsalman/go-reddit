# GO-REDDIT COLLECTOR

## Import
```
go get github.com/soumitsalman/go-reddit
```

## Sample Code
```
import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	ds "github.com/soumitsalman/beansack/sdk"
	"github.com/soumitsalman/go-reddit/collector"
)

func collectAndStore() {
	config :=collector.NewCollectorConfig(localFileStore)
	collector.NewCollector(config).Collect()
}

func localFileStore(contents []ds.Bean) {
	filename := fmt.Sprintf("outputs_REDDIT_%s", time.Now().Format("2006-01-02-15-04-05.json"))
	file, _ := os.Create(filename)
	defer file.Close()
	json.NewEncoder(file).Encode(contents)

}

```

This has the most common read functions for reddit api