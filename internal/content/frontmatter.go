package content

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

type FrontMatter struct {
	Title string `toml:"title"`
	Date  string `toml:"date"`
	Draft bool   `toml:"draft"`
}

// ParseFrontMatter extracts TOML metadata and returns the remaining Markdown body.
func ParseFrontMatter(source []byte) (FrontMatter, []byte, error) {
	var fm FrontMatter
	var body []byte

	if !bytes.HasPrefix(source, []byte("+++\n")) && !bytes.HasPrefix(source, []byte("+++\r\n")) {
		return fm, body, errors.New("front matter is missing")
	}

	normalized := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))

	rest := normalized[len("+++\n"):]

	end := bytes.Index(rest, []byte("\n+++\n"))
	if end < 0 {
		if bytes.HasSuffix(rest, []byte("\n+++")) {
			end = len(rest) - len("\n+++")
			fm, err := parseTOMLFrontMatter(rest[:end])
			if err != nil {
				return fm, body, err
			}
			return fm, body, nil
		}

		return fm, body, errors.New("front matter is not closed")
	}

	fm, err := parseTOMLFrontMatter(rest[:end])
	if err != nil {
		return fm, body, err
	}

	body = rest[end+len("\n+++\n"):]

	return fm, body, nil
}

// parseTOMLFrontMatter decodes TOML metadata and wraps any parsing error.
func parseTOMLFrontMatter(source []byte) (FrontMatter, error) {
	var fm FrontMatter
	if err := toml.Unmarshal(source, &fm); err != nil {
		return fm, fmt.Errorf("failed to parse front matter TOML: %w", err)
	}
	return fm, nil
}
