package main

import (
	"github.com/ari-kl/phish-stream/config"
	"github.com/ari-kl/phish-stream/filter"
	"github.com/ari-kl/phish-stream/review"
	"github.com/ari-kl/phish-stream/util"
)

func main() {
	config.LoadConfig()
	util.SetupLogger()
	filter.InitFilters()
	go review.StartSlackBot()
	StartStreaming()
}
