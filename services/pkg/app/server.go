package app

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

type GetFeedSkeletonResponse struct {
	Feed []Post `json:"feed"`
}

type Post struct {
	Post string `json:"post"`
}

func Server() error {
	slog.Info("starting server")

	app, err := NewApp()
	if err != nil {
		return util.WrapErr("failed to create app", err)
	}
	defer app.Close()

	server := http.NewServeMux()

	// API status endpoint
	server.HandleFunc("/xrpc/app.bsky.feed.getFeedSkeleton", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public; max-age=10")

		// Fetch post records from the cache
		posts, _, err := app.Cache.FindPosts(0)
		if err != nil {
			slog.Error("failed to find posts", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Converto to response format
		feed := make([]Post, len(posts))
		for i, post := range posts {
			feed[i] = Post{
				Post: post.AtURI,
			}
		}
		resp := GetFeedSkeletonResponse{
			Feed: feed,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("failed to encode response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	slog.Info("starting server on :8080")
	return http.ListenAndServe(":8080", server)
}
