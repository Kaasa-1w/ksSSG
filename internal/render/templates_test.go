package render

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ks-ssg/internal/content"
)

// TestRenderSelectsTemplateByTermStatus verifies both node-specific templates.
func TestRenderSelectsTemplateByTermStatus(t *testing.T) {
	renderer, err := NewRenderer("../../test/layout")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		isTermNode    bool
		expectedClass string
	}{
		{name: "term", isTermNode: true, expectedClass: `class="term"`},
		{name: "non-term", isTermNode: false, expectedClass: `class="non-term"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer

			err := renderer.Render(&output, &content.Node{
				IsTermNode: test.isTermNode,
				Title:      "Example",
				Content:    template.HTML("<p>body</p>"),
			})
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(output.String(), test.expectedClass) {
				t.Fatalf("output does not contain %q:\n%s", test.expectedClass, output.String())
			}
			if !strings.Contains(output.String(), "<title>Example</title>") {
				t.Fatalf("output does not contain the shared base template:\n%s", output.String())
			}
			if !strings.Contains(output.String(), "<p>body</p>") {
				t.Fatalf("output does not contain rendered content:\n%s", output.String())
			}
		})
	}
}

func TestNewRendererLoadsAddedTemplatesAndCrossFileDefinitions(t *testing.T) {
	layoutDir := t.TempDir()
	writeTemplateFile(t, layoutDir, "base.html", `{{template "card" .}}`)
	writeTemplateFile(t, layoutDir, "components/card.html", `{{define "card"}}Card: {{template "sample" .}}{{end}}`)
	writeTemplateFile(t, layoutDir, "components/sample.html", `{{define "sample"}}{{.Title}}{{end}}`)

	renderer, err := NewRenderer(layoutDir)
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	var output bytes.Buffer
	if err := renderer.Render(&output, &content.Node{Title: "Example"}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got, want := output.String(), "Card: Example"; got != want {
		t.Errorf("rendered output = %q, want %q", got, want)
	}
}

func writeTemplateFile(t *testing.T, root, relativePath, contents string) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}
