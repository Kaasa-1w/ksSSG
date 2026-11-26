package render

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"ks-ssg/internal/content"
)

type Renderer struct {
	templates *template.Template
}

// NewRenderer parses the layout templates and returns a reusable renderer.
func NewRenderer(layoutDir string) (*Renderer, error) {
	var templateFiles []string
	err := filepath.WalkDir(layoutDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to walk layout directory at %q: %w", path, walkErr)
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}

		relativePath, err := filepath.Rel(layoutDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for layout template %q: %w", path, err)
		}
		templateFiles = append(templateFiles, filepath.ToSlash(relativePath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find layout templates in %q: %w", layoutDir, err)
	}
	if len(templateFiles) == 0 {
		return nil, fmt.Errorf("no HTML templates found in layout directory %q", layoutDir)
	}

	tmpl, err := template.ParseFS(os.DirFS(layoutDir), templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout templates in %q: %w", layoutDir, err)
	}
	if tmpl.Lookup("base.html") == nil {
		return nil, fmt.Errorf("base.html template not found in layout directory %q", layoutDir)
	}

	return &Renderer{templates: tmpl}, nil
}

// Render executes the shared page template and its node-specific content template.
func (r *Renderer) Render(writer io.Writer, node *content.Node) error {
	return r.templates.ExecuteTemplate(writer, "base.html", node)
}
