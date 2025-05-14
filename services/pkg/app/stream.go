package app

import (
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

// StreamEvent (and subtypes) represent a message from the Jetstream.
// Fields for both posts and reposts are included.
type StreamEvent struct {
	DID    string `json:"did"`
	TimeUS int64  `json:"time_us"`
	Kind   string `json:"kind"`
	Commit Commit `json:"commit"`
}

type Commit struct {
	Operation string `json:"operation"`
	Record    Record `json:"record"`
	RKey      string `json:"rkey"`
	CID       string `json:"cid"`
}

type Record struct {
	Type      string   `json:"$type"`
	Languages []string `json:"langs"`
	Embed     Embed    `json:"embed"`
	Facets    []Facet  `json:"facets"`
	Subject   Content  `json:"subject"`
	Reply     Reply    `json:"reply"`
}

type Embed struct {
	Type     string        `json:"$type"`
	External ExternalEmbed `json:"external"`
	Record   RecordEmbed   `json:"record"`
}

type ExternalEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	Thumb       Thumb  `json:"thumb"`
}

type RecordEmbed struct {
	CID string `json:"cid"`
	URI string `json:"uri"`
}

type Thumb struct {
	Type     string `json:"$type"`
	Ref      Ref    `json:"ref"`
	MimeType string `json:"mimeType"`
}

type Ref struct {
	Link string `json:"$link"`
}

type Facet struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	Type string `json:"$type"`
	URI  string `json:"uri"`
}

type Reply struct {
	Parent Content `json:"parent"`
	Root   Content `json:"root"`
}

type Content struct {
	CID string `json:"cid"`
	URI string `json:"uri"`
}

// Valid determines whether a stream event should be processed by our application.
func (s *StreamEvent) Valid() bool {
	if s.Kind != "commit" {
		return false
	}
	if s.Commit.Operation != "create" {
		return false
	}
	if !s.IsPost() && !s.IsRepost() && !s.IsLike() {
		return false
	}
	if s.IsPost() && !s.IsEnglish() {
		return false
	}

	return true
}

func (s *StreamEvent) IsPost() bool {
	return s.Commit.Record.Type == "app.bsky.feed.post"
}

// IsStandardPost determines whether the post is 'standard', i.e. not a quote or reply.
// We also exclude posts with link facets, as we are trying to avoid 'commentary' on news.
func (s *StreamEvent) IsStandardPost() bool {
	return s.IsPost() && !s.IsReplyPost() && !s.HasEmbed() && !s.HasFacet()
}

// IsQuotePost determines whether the event is a quote post.
// Quote posts are a subset of posts that contain record embeds.
func (s *StreamEvent) IsQuotePost() bool {
	if !s.IsPost() {
		return false
	}
	return s.Commit.Record.Embed.Type == "app.bsky.embed.record"
}

func (s *StreamEvent) IsReplyPost() bool {
	if !s.IsPost() {
		return false
	}
	return s.Commit.Record.Reply.Parent.CID != "" && s.Commit.Record.Reply.Parent.URI != ""
}

func (s *StreamEvent) IsRepost() bool {
	return s.Commit.Record.Type == "app.bsky.feed.repost"
}

func (s *StreamEvent) IsLike() bool {
	return s.Commit.Record.Type == "app.bsky.feed.like"
}

func (s *StreamEvent) IsEnglish() bool {
	return util.ContainsStr(s.Commit.Record.Languages, "en")
}

func (s *StreamEvent) HasFacet() bool {
	return len(s.Commit.Record.Facets) > 0
}

func (s *StreamEvent) HasEmbed() bool {
	return s.Commit.Record.Embed.Type != ""
}
