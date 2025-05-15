package app

import (
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

var badWords = []string{
	"trump",
	"trump's",
	"biden",
	"biden's",
	"putin",
	"rfk",
	"elon",
	"musk",
	"schumer",
	"pelosi",
	"maga",
	"republican",
	"republicans",
	"democrat",
	"democrats",
	"gop",
	"genocide",
	"wordle",
}

// Posts containing blocked words are excluded from the feed.
// This is to prevent the bulk of junk, spam, and political posts.
func blockedWords() mapset.Set[string] {
	set := mapset.NewSet[string]()
	for _, word := range badWords {
		set.Add(word)
	}
	return set
}

// Posts from blocked DIDs are excluded from the feed.
// DIDs in this list are known bots.
func blockedDIDs() mapset.Set[string] {
	set := mapset.NewSet[string]()

	// Read the embedded file dids.txt file
	data, err := assets.ReadFile("assets/dids.txt")
	if err != nil {
		slog.Error(util.WrapErr("failed to read dids.txt", err).Error())
		return set
	}

	// Split the data into lines and add each line to the set.
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		set.Add(line)
	}

	return set
}
