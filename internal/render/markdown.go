package render

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
)

// MarkdownToHTML converts Markdown source into trusted HTML.
func MarkdownToHTML(source []byte) (template.HTML, error) {
	var buf bytes.Buffer

	if err := goldmark.Convert(source, &buf); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}
