package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

func Server() error {
	slog.Info("starting server")

	app, err := NewApp()
	if err != nil {
		return util.WrapErr("failed to create app", err)
	}
	defer app.Close()

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
		w.Header().Set("Cache-Control", "public; max-age=10")

		// Handle "limit" query parameter
		limitStr := r.URL.Query().Get("limit")
		if limitStr == "" {
			limitStr = "10"
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			slog.Error("failed to parse limit", "error", err)
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		if limit < 1 || limit > 100 {
			slog.Error("invalid limit", "limit", limit)
			http.Error(w, "Limit must be between 1 and 100", http.StatusBadRequest)
			return
		}

		// Look for "cursor" query parameter
		cursorStr := r.URL.Query().Get("cursor")
		if cursorStr == "" {
			cursorStr = "0"
		}
		cursor, err := strconv.ParseUint(cursorStr, 10, 64)
		if err != nil {
			slog.Error("failed to parse cursor", "error", err)
			http.Error(w, "Invalid cursor", http.StatusBadRequest)
			return
		}
		if cursor < 0 {
			slog.Error("invalid cursor", "cursor", cursor)
			http.Error(w, "Invalid cursor", http.StatusBadRequest)
			return
		}

		// Fetch post records from the cache
		posts, respCursor, err := app.Cache.FindPosts(limit, cursor)
		if err != nil {
			slog.Error("failed to find posts", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Convert to response fomat
		feed := make([]APIPost, len(posts))
		for i, post := range posts {
			feed[i] = APIPost{
				Post: post.AtURI,
			}
		}
		respCursorStr := strconv.FormatUint(respCursor, 10)
		if respCursorStr == "0" {
			respCursorStr = "" // Ensure cursor is omitted from response if no more posts are available
		}
		resp := APIFeedSkeletonResponse{
			Feed:   feed,
			Cursor: respCursorStr,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("failed to encode response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	slog.Info("starting server", "port", app.Config.ServerPort)
	return http.ListenAndServe(fmt.Sprintf(":%s", app.Config.ServerPort), server)
}
