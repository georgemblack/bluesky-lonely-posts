package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/cache"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

func Server() error {
	slog.Info("starting server")

	app, err := NewApp()
	if err != nil {
		return util.WrapErr("failed to create app", err)
	}
	defer app.Close()

	// Update Cloudflare DNS records
	if err := updateServiceDNS(app.Config); err != nil {
		slog.Error(util.WrapErr("failed to update dns", err).Error())
	}

	server := http.NewServeMux()

	// Serve the DID document for this domain.
	server.HandleFunc("/.well-known/did.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public; max-age=28800") // 8 hours

		didDoc := APIDIDDocResponse{
			Context: []string{"https://www.w3.org/ns/did/v1"},
			ID:      "did:web:feedgen.george.black",
			Service: []APIService{{
				ID:              "#bsky_fg",
				Type:            "BskyFeedGenerator",
				ServiceEndpoint: "https://feedgen.george.black",
			}},
		}

		if err := json.NewEncoder(w).Encode(didDoc); err != nil {
			slog.Error("failed to encode response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// Serve the feed by supplying a list of AT URIs of "lonely posts".
	// Read a random set of ten posts from the cache.
	server.HandleFunc("/xrpc/app.bsky.feed.getFeedSkeleton", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public; max-age=15")

		slog.Info("request", "limit", r.URL.Query().Get("limit"), "cursor", r.URL.Query().Get("cursor"))

		// Fetch post records from the cache
		posts, respCursor, err := app.Cache.FindPosts(limitQuery(r), cursorQuery(r))
		if err != nil {
			slog.Error("failed to find posts", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Format & encode response
		resp := toResponse(posts, respCursor)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("failed to encode response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	slog.Info("starting server", "port", app.Config.ServerPort)
	return http.ListenAndServe(fmt.Sprintf(":%s", app.Config.ServerPort), server)
}

func limitQuery(r *http.Request) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 10
	}
	if limit < 1 || limit > 100 {
		return 10
	}

	return limit
}

func cursorQuery(r *http.Request) uint64 {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}
	cursor, err := strconv.ParseUint(cursorStr, 10, 64)
	if err != nil {
		return 0
	}
	if cursor < 0 {
		return 0
	}

	return cursor
}

func toResponse(posts []cache.PostRecord, cursor uint64) APIFeedSkeletonResponse {
	feed := make([]APIPost, len(posts))
	for i, post := range posts {
		feed[i] = APIPost{
			Post: post.AtURI,
		}
	}
	cursorStr := strconv.FormatUint(cursor, 10)
	if cursorStr == "0" || len(posts) == 0 {
		cursorStr = "" // Ensure cursor is omitted from response if no more posts are available
	}

	return APIFeedSkeletonResponse{
		Feed:   feed,
		Cursor: cursorStr,
	}
}
