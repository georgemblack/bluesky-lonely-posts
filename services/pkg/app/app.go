package app

import (
	"embed"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/cache"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/config"
)

//go:embed assets/*
var assets embed.FS

// App creates a new instance of the application, initializing the cache, storage, and Bluesky API client.
type App struct {
	Config config.Config
	Cache  Cache
}

func NewApp() (App, error) {
	config, err := config.New()
	if err != nil {
		return App{}, err
	}

	cache, err := cache.New(config)
	if err != nil {
		return App{}, err
	}

	return App{
		Config: config,
		Cache:  cache,
	}, nil
}

func (a App) Close() {
	a.Cache.Close()
}
