package main

import (
	"github.com/soumitsalman/goreddit/api"
)

func main() {
	api.NewServer(2, 5).Run()
}
