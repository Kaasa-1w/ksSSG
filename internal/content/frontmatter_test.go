package content

import (
	"strings"
	"testing"
)

// TestParseFrontMatter covers valid metadata and malformed front matter inputs.
func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantTitle   string
		wantDraft   bool
		wantBody    string
		wantErrText string
	}{
		{
			name:      "metadata and body",
			source:    "+++\ntitle = \"Page\"\ndraft = true\n+++\nbody",
			wantTitle: "Page",
			wantDraft: true,
			wantBody:  "body",
		},
		{
			name:      "CRLF",
			source:    "+++\r\ntitle = \"Page\"\r\n+++\r\nbody",
			wantTitle: "Page",
			wantBody:  "body",
		},
		{
			name:      "empty body",
			source:    "+++\ntitle = \"Page\"\n+++",
			wantTitle: "Page",
		},
		{
			name:      "comments and escaped string",
			source:    "+++\n# page metadata\ntitle = \"A \\\"quoted\\\" page\" # title\n date = '2026/10/30'\n draft = false\n+++\nbody",
			wantTitle: "A \"quoted\" page",
			wantBody:  "body",
		},
		{
			name:        "missing front matter",
			source:      "body",
			wantErrText: "front matter is missing",
		},
		{
			name:        "unclosed front matter",
			source:      "+++\ntitle = \"Page\"\nbody",
			wantErrText: "front matter is not closed",
		},
		{
			name:        "invalid TOML",
			source:      "+++\ntitle = [\n+++\nbody",
			wantErrText: "failed to parse front matter TOML",
		},
		{
			name:        "duplicate key",
			source:      "+++\ntitle = \"First\"\ntitle = \"Second\"\n+++\nbody",
			wantErrText: "duplicate key",
		},
		{
			name:        "wrong field type",
			source:      "+++\ndraft = \"yes\"\n+++\nbody",
			wantErrText: "draft must be true or false",
		},
		{
			name:        "non-TOML boolean",
			source:      "+++\ndraft = TRUE\n+++\nbody",
			wantErrText: "draft must be true or false",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fm, body, err := ParseFrontMatter([]byte(test.source))
			if test.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErrText) {
					t.Fatalf("error = %v, want text %q", err, test.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFrontMatter() error = %v", err)
			}
			if fm.Title != test.wantTitle {
				t.Errorf("Title = %q, want %q", fm.Title, test.wantTitle)
			}
			if fm.Draft != test.wantDraft {
				t.Errorf("Draft = %t, want %t", fm.Draft, test.wantDraft)
			}
			if string(body) != test.wantBody {
				t.Errorf("body = %q, want %q", body, test.wantBody)
			}
		})
	}
}

// TestParseTOMLFieldsSupportsAdditionalStructFields verifies that tagged fields are decoded without key-specific parser logic.
func TestParseTOMLFieldsSupportsAdditionalStructFields(t *testing.T) {
	type extendedFrontMatter struct {
		Author   string `toml:"author"`
		Featured bool   `toml:"featured"`
		Rank     int    `toml:"rank"`
	}

	var metadata extendedFrontMatter
	err := parseTOMLFields([]byte("author = \"Ada\"\nfeatured = true\nrank = 3"), &metadata)
	if err != nil {
		t.Fatalf("parseTOMLFields() error = %v", err)
	}
	if metadata.Author != "Ada" || !metadata.Featured || metadata.Rank != 3 {
		t.Errorf("metadata = %+v, want author Ada, featured true, rank 3", metadata)
	}
}
