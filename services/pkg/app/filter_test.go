package app

import "testing"

const (
	validDID   = "did:plc:example"
	blockedDID = "did:plc:zrcqicmkxum6tir6ahthppif" // @theblue.report
)

func TestIncludePost(t *testing.T) {
	tests := []struct {
		name     string
		input    StreamEvent
		expected bool
	}{
		{
			name:     "include valid post",
			input:    streamEvent("woohoo!", validDID),
			expected: true,
		},
		{
			name:     "exclude posts with empty text",
			input:    streamEvent("", validDID),
			expected: false,
		},
		{
			name:     "exclude posts with blocked DID",
			input:    streamEvent("valid post", blockedDID),
			expected: false,
		},
		{
			name:     "exclude post in all caps",
			input:    streamEvent("WHY. IS. EVERYONE. SHOUTING!!!", validDID),
			expected: false,
		},
		{
			name:     "exclude post with banned word",
			input:    streamEvent("here's what i think about that TRUMP guy!", validDID),
			expected: false,
		},
		{
			name:     "exclude post mistakenly labeled as English (1)",
			input:    streamEvent("يرجى تجاهل المنشور التجريبي", validDID),
			expected: false,
		},
		{
			name:     "exclude post mistakenly labeled as English (2)",
			input:    streamEvent("测试帖子请忽略", validDID),
			expected: false,
		},
		{
			name:     "exclude 'Connections' game post",
			input:    streamEvent("Connections Puzzle #15", validDID),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := includePost(test.input)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func streamEvent(text, did string) StreamEvent {
	return StreamEvent{
		DID: did,
		Commit: Commit{
			Record: Record{
				Text: text,
			},
		},
	}
}
