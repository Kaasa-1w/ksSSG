package render

import (
	"html/template"
	"io"
	"ks-ssg/internal/content"
	"path/filepath"
)

type Renderer struct {
	templates *template.Template
}

// NewRenderer parses the layout templates and returns a reusable renderer.
func NewRenderer(layoutDir string) (*Renderer, error) {
	tmpl, err := template.ParseFiles(
		filepath.Join(layoutDir, "base.html"),
		filepath.Join(layoutDir, "term.html"),
		filepath.Join(layoutDir, "non-term.html"),
	)
	if err != nil {
		return nil, err
	}

	return &Renderer{templates: tmpl}, nil
}

// Render executes the shared page template and its node-specific content template.
func (r *Renderer) Render(writer io.Writer, node *content.Node) error {
	return r.templates.ExecuteTemplate(writer, "base.html", node)
}
