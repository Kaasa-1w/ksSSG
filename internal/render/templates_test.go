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

			node := &content.Node{
				IsTermNode: test.isTermNode,
				Title:      "Example",
				Content:    template.HTML("<p>body</p>"),
			}
			err := renderer.Render(&output, node)
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
			if got, want := node.Content, template.HTML("<p>body</p>"); got != want {
				t.Errorf("Render() changed node content to %q, want %q", got, want)
			}
		})
	}
}

func TestNewRendererLoadsAddedTemplatesAndCrossFileDefinitions(t *testing.T) {
	layoutDir := t.TempDir()
	writeTemplateFile(t, layoutDir, "base.html", `{{.Content}}`)
	writeTemplateFile(t, layoutDir, "term.html", `{{define "term"}}{{template "card" .}}{{end}}`)
	writeTemplateFile(t, layoutDir, "non-term.html", `{{define "non-term"}}{{template "card" .}}{{end}}`)
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

func TestRenderMarkdownExecutesShortcodeTemplatesWithPageAndArguments(t *testing.T) {
	layoutDir := t.TempDir()
	writeTemplateFile(t, layoutDir, "base.html", `{{.Content}}`)
	writeTemplateFile(t, layoutDir, "shortcodes/link.html", `{{define "shortcode:link"}}{{if eq .Args.route "about"}}<a href="/about.html">{{else}}<a href="{{.Args.url}}">{{end}}{{if .Args.label}}{{.Args.label}}{{else}}{{.Page.Title}}{{end}}</a>{{end}}`)
	writeTemplateFile(t, layoutDir, "shortcodes/wrapper.html", `{{define "shortcode:wrapper"}}<section>{{template "shortcode:link" .}}</section>{{end}}`)

	renderer, err := NewRenderer(layoutDir)
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	got, err := renderer.RenderMarkdown(
		[]byte(`See {{< wrapper url="/about?x=a&y=b" label="About & info" >}}.`),
		&content.Node{Title: "Fallback"},
	)
	if err != nil {
		t.Fatalf("RenderMarkdown() error = %v", err)
	}
	if want := `<p>See <section><a href="/about?x=a&amp;y=b">About &amp; info</a></section>.</p>` + "\n"; string(got) != want {
		t.Errorf("RenderMarkdown() = %q, want %q", got, want)
	}

	got, err = renderer.RenderMarkdown([]byte(`{{< link route="about" >}}`), &content.Node{Title: "Fallback"})
	if err != nil {
		t.Fatalf("RenderMarkdown() without label error = %v", err)
	}
	if want := `<a href="/about.html">Fallback</a>` + "\n"; string(got) != want {
		t.Errorf("RenderMarkdown() without label = %q, want %q", got, want)
	}
}

func TestRenderMarkdownReportsUnknownAndFailingShortcodes(t *testing.T) {
	layoutDir := t.TempDir()
	writeTemplateFile(t, layoutDir, "base.html", `{{.Content}}`)
	writeTemplateFile(t, layoutDir, "shortcodes/failure.html", `{{define "shortcode:failure"}}{{.Page.Missing}}{{end}}`)

	renderer, err := NewRenderer(layoutDir)
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	for _, test := range []struct {
		name     string
		markdown string
		want     string
	}{
		{name: "unknown", markdown: `{{< missing >}}`, want: `template "shortcode:missing" is not defined`},
		{name: "execution failure", markdown: `{{< failure >}}`, want: `shortcode "failure"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderer.RenderMarkdown([]byte(test.markdown), &content.Node{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("RenderMarkdown() error = %v, want error containing %q", err, test.want)
			}
		})
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
