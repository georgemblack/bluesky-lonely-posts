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
	"quordle",
}

// Get a set of blocked DIDs that can be used to filter out bot posts.
func getBlockedDIDs() mapset.Set[string] {
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

// Get a set of blocked words that can be used to filter out unwanted posts.
func getBlockedWords() mapset.Set[string] {
	set := mapset.NewSet[string]()
	for _, word := range badWords {
		set.Add(word)
	}
	return set
}

// Get a set of all lowercase a-z characters.
// This is used to check for at least one lowercase character in the post text.
func getAToZ() mapset.Set[string] {
	set := mapset.NewSet[string]()
	for i := 'a'; i <= 'z'; i++ {
		set.Add(string(i))
	}
	return set
}

var blockedDIDs = getBlockedDIDs()
var blockedWords = getBlockedWords()
var aToZ = getAToZ()

// Given the contents of a post, and the DID of the user who posted it, determine if the post should be included in the feed.
// Perform the following checks:
//  1. Check for blocked DIDs (bots)
//  2. Check for blocked words in post text
//  3. Check to ensure post has at least one lowercase a-z character
//     - This prevents ALL CAPS SHOUTING POSTS, as well as posts mistakenly labeled as English
func includePost(event StreamEvent) bool {
	did := event.GetDID()
	text := event.GetText()

	if text == "" {
		return false
	}

	// Check for blocked DIDs (bots)
	if blockedDIDs.Contains(did) {
		return false
	}

	// Check for blocked words in post text
	tokens := strings.Fields(text)
	for _, token := range tokens {
		if blockedWords.Contains(strings.ToLower(token)) {
			return false
		}
	}

	// Check for the words 'Connections' and 'Puzzle' in post text (case sensitive).
	// This is used to filter out a set of posts similar to Wordle posts.
	// Directly banning the world 'connections' will not work, as it is used in other contexts.
	if strings.Contains(text, "Connections") && strings.Contains(text, "Puzzle") {
		return false
	}

	// Check for at least one lowercase a-z character
	lowercase := false
	for _, char := range text {
		if aToZ.Contains(string(char)) {
			lowercase = true
			break
		}
	}
	if !lowercase {
		return false
	}

	return true
}
