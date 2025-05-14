package app

import (
	"github.com/georgemblack/bluesky-lonely-posts/pkg/cache"
)

type Cache interface {
	SavePost(hash string, post cache.PostRecord) error
	ReadPost(hash string) (cache.PostRecord, error)
	DeletePost(hash string) error
	FindPosts(cursor uint64) ([]cache.PostRecord, uint64, error)
	Close()
}
