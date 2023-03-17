package main

import (
	"github.com/lleadbet/chat-controller/twitch"
	"go.uber.org/zap"
)

func main() {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	logger := zap.Must(config.Build()).WithOptions(zap.WithCaller(false))
	defer logger.Sync()

	logger.Info("Starting up chat controller...")

	twitch.ChatReader(logger)
}
