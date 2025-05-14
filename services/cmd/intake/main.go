package main

import (
	"log/slog"
	"os"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/app"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	err := app.Intake()
	if err != nil {
		slog.Error(err.Error())
	}
}
