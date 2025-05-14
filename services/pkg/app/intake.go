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
	saves     int
	ignored   int
	deletions int
}

func newStats() Stats {
	return Stats{
		started:   time.Now(),
		errors:    0,
		saves:     0,
		ignored:   0,
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

	// Start worker threads
	var wg sync.WaitGroup
	wg.Add(WorkerPoolSize)
	stream := make(chan StreamEvent, StreamBufferSize)
	shutdown := make(chan struct{})
	for i := 0; i < WorkerPoolSize; i++ {
		go intakeWorker(i+1, stream, shutdown, app, &wg)
	}

	// Connect to Jetstream
	conn, _, err := websocket.DefaultDialer.Dial(JetstreamURL, nil)
	if err != nil {
		return util.WrapErr("failed to dial jetstream", err)
	}
	defer conn.Close()

	// Send Jetstream messages to workers
	errors := 0
	for {
		event := StreamEvent{}
		err := conn.ReadJSON(&event)
		if err != nil {
			errors++
			slog.Warn(util.WrapErr("failed to read json", err).Error())

			if errors > ErrorThreshold {
				slog.Error("encountered too many errors reading from jetstream")
				break
			}

			continue
		}

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

		if !event.Valid() {
			stats.ignored++
			continue
		}

		if event.IsStandardPost() {
			// Save standard posts to the cache.
			// Standard posts don't interact with other posts (i.e. quotes & replies), and don't contain external links.
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
			// For other event types, find the referenced post (if applicable) and delete it from the cache.
			// This includes likes, reposts, replies, and quotes.
			atURI := referencedPost(event)
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
			slog.Info("intake stats", "errors", stats.errors, "saves", stats.saves, "deletions", stats.deletions, "ignored", stats.ignored, "queue", len(stream))
			stats = newStats()
		}
	}
}

// Given a stream event that references a post, return the AT URI of the post it is referencing.
func referencedPost(event StreamEvent) string {
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
