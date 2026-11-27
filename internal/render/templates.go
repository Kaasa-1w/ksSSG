package render

import (
	"bytes"
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

// Render selects a node-specific content template and renders it inside the shared page template.
func (r *Renderer) Render(writer io.Writer, node *content.Node) error {
	contentTemplate := "term"
	if !node.IsTermNode {
		contentTemplate = "non-term"
	}

	var contentOutput bytes.Buffer
	if err := r.templates.ExecuteTemplate(&contentOutput, contentTemplate, node); err != nil {
		return fmt.Errorf("execute %q template: %w", contentTemplate, err)
	}

	renderedNode := *node
	renderedNode.Content = template.HTML(contentOutput.String())
	if err := r.templates.ExecuteTemplate(writer, "base.html", &renderedNode); err != nil {
		return fmt.Errorf("execute base.html template: %w", err)
	}
	return nil
}

// RenderMarkdown converts Markdown and executes user-defined shortcode templates.
func (r *Renderer) RenderMarkdown(source []byte, page *content.Node) (template.HTML, error) {
	return markdownToHTML(source, func(name string, args map[string]string) (template.HTML, error) {
		templateName := "shortcode:" + name
		if r.templates.Lookup(templateName) == nil {
			return "", fmt.Errorf("template %q is not defined", templateName)
		}

		var output bytes.Buffer
		data := struct {
			Page *content.Node
			Args map[string]string
		}{
			Page: page,
			Args: args,
		}
		if err := r.templates.ExecuteTemplate(&output, templateName, data); err != nil {
			return "", fmt.Errorf("execute template %q: %w", templateName, err)
		}
		return template.HTML(output.String()), nil
	})
}
