package app

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/cache"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
	"github.com/gorilla/websocket"
)

const (
	WorkerPoolSize   = 1
	StreamBufferSize = 10000
	ErrorThreshold   = 10
	JetstreamURL     = "wss://jetstream2.us-east.bsky.network/subscribe?wantedCollections=app.bsky.feed.post&wantedCollections=app.bsky.feed.repost&wantedCollections=app.bsky.feed.like"
)

type Stats struct {
	started   time.Time
	errors    int
	ignored   int // Number ignored events
	saves     int // Number of posts saved to the cache
	blocked   int // Number of posts blocked by filters
	deletions int // Number of deletions from the cache
}

func newStats() Stats {
	return Stats{
		started:   time.Now(),
		errors:    0,
		ignored:   0,
		saves:     0,
		deletions: 0,
	}
}

func Intake() error {
	slog.Info("starting intake")

	app, err := NewApp()
	if err != nil {
		return util.WrapErr("failed to create app", err)
	}
	defer app.Close()

	// Start worker threads.
	// Each worker thread reads from the queue of events and processes them.
	var wg sync.WaitGroup
	wg.Add(WorkerPoolSize)
	stream := make(chan StreamEvent, StreamBufferSize)
	shutdown := make(chan struct{})
	for i := 0; i < WorkerPoolSize; i++ {
		go intakeWorker(i+1, stream, shutdown, app, &wg)
	}

	// Read the Jetstream cursor from the cache.
	// If our application exited due to an error, our position in the Jetstream may have been saved.
	cursor, err := app.Cache.ReadCursor()
	if err != nil {
		slog.Warn(util.WrapErr("failed to read cursor", err).Error())
	} else {
		if cursor > 0 {
			slog.Info("discovered cursor", "cursor", cursor)

			// Subtrack 5 seconds from the cursor to ensure we don't miss events
			cursor -= 5 * 1_000_000

			slog.Info("using cursor", "cursor", cursor)
		} else {
			slog.Info("no cursor found, continuing without")
		}
	}

	// Build the URL to connect to the Jetstream
	url := JetstreamURL
	if cursor > 0 {
		url += fmt.Sprintf("&cursor=%d", cursor)
	}

	// Connect to the Jetstream
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return util.WrapErr("failed to dial jetstream", err)
	}
	defer conn.Close()

	// Keep track of the most recent event timestamp.
	// If our connection is interrupted, we can use this to resume from the last event.
	var latest int64

	// Parse Jetstream messages as JSON and send them to the worker queue
	errors := 0
	for {
		event := StreamEvent{}
		err := conn.ReadJSON(&event)

		if err != nil {
			errors++
			slog.Warn(util.WrapErr("failed to read json", err).Error())

			// If we encounter too many errors, save our position in the Jetstream and exit
			if errors > ErrorThreshold {
				slog.Error("encountered too many errors reading from jetstream, saving cursor and exiting")

				if latest > 0 {
					err := app.Cache.SaveCursor(latest)
					if err != nil {
						slog.Error(util.WrapErr("failed to save cursor", err).Error())
					} else {
						slog.Info("saved cursor", "cursor", latest)
					}
				} else {
					slog.Warn("no cursor to save, exiting")
				}

				break
			}

			continue
		}

		latest = event.TimeUS
		stream <- event
	}

	// Signal workers to exit, and wait for them to finish
	close(shutdown)
	wg.Wait()
	return nil
}

func intakeWorker(id int, stream chan StreamEvent, shutdown chan struct{}, app App, wg *sync.WaitGroup) {
	slog.Info(fmt.Sprintf("starting worker %d", id))
	defer wg.Done()

	stats := newStats()

	for {
		event := StreamEvent{}
		ok := true

		select {
		case event, ok = <-stream:
			if !ok {
				slog.Error("error reading message from channel")
				continue
			}
		case <-shutdown:
			slog.Info(fmt.Sprintf("shutting down worker %d", id))
			return
		}

		// Determine whether the event should be processed
		if !event.Valid() {
			stats.ignored++
			continue
		}

		// Process the event
		if event.IsStandardPost() {
			// Standard posts are standalone posts (i.e. not quotes, replies) and don't contain any media or external links.
			// These posts are elligible to appear in the feed.

			// Determine whether the post passes content filters.
			if !includePost(event) {
				stats.blocked++
				continue
			}

			// Save to cache in order for it to be displayed in the feed.
			atURI := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", event.DID, event.Commit.RKey)
			err := app.Cache.SavePost(util.Hash(atURI), cache.PostRecord{
				AtURI:     atURI,
				Timestamp: event.TimeUS,
			})
			if err != nil {
				slog.Error(fmt.Sprintf("failed to save post %s", atURI), "error", err)
				stats.errors++
				continue
			}
			stats.saves++
		} else {
			// For all other events, determine if they interact with a post (i.e. likes, quotes, replies).
			// Delete the target post from the cache if it exists, to prevent it from appearing in the feed.
			atURI := targetPost(event)
			if atURI == "" {
				stats.ignored++
				continue
			}
			if err := app.Cache.DeletePost(util.Hash(atURI)); err != nil {
				slog.Error(util.WrapErr("failed to delete post", err).Error(), "at_uri", atURI)
				stats.errors++
				continue
			}
			stats.deletions++
		}

		// Log stats every ~5 minutes
		if time.Since(stats.started) > 5*time.Minute {
			slog.Info("intake stats", "saves", stats.saves, "deletions", stats.deletions, "errors", stats.errors, "ignored", stats.ignored, "blocked", stats.blocked, "queue", len(stream))
			stats = newStats()
		}
	}
}

// Given a stream event that references a post, return the AT URI of the post it is referencing.
func targetPost(event StreamEvent) string {
	if event.IsLike() || event.IsRepost() {
		return event.Commit.Record.Subject.URI
	}
	if event.IsQuotePost() {
		return event.Commit.Record.Embed.Record.URI
	}
	if event.IsReplyPost() {
		return event.Commit.Record.Reply.Parent.URI
	}
	return ""
}
