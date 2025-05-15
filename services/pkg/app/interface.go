package app

import (
	"github.com/georgemblack/bluesky-lonely-posts/pkg/cache"
)

type Cache interface {
	SavePost(hash string, post cache.PostRecord) error
	ReadPost(hash string) (cache.PostRecord, error)
	ReadPosts(n int, cursor uint64) ([]cache.PostRecord, uint64, error)
	DeletePost(hash string) error
	SaveCursor(cursor int64) error
	ReadCursor() (int64, error)
	Close()
}
