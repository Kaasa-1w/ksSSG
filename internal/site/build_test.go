package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ks-ssg/internal/content"
)

// TestBuildRendersPublishedPagesAndSkipsDrafts checks output generation and draft exclusion.
func TestBuildRendersPublishedPagesAndSkipsDrafts(t *testing.T) {
	root := t.TempDir()
	config := writeSiteFixture(t, root)
	writeFile(t, config.SourceDir, "index.md", page("Home", false, "# Home"))
	writeFile(t, config.SourceDir, "about.md", page("About", true, "# About"))
	writeFile(t, config.SourceDir, "articles/index.md", page("Articles", false, "# Articles"))
	writeFile(t, config.SourceDir, "articles/first.md", page("First", false, "# First"))
	writeFile(t, config.SourceDir, "articles/reference/syntax.md", page("Syntax", false, "# Syntax"))
	writeFile(t, config.SourceDir, "ignored.txt", "not markdown")

	if err := Build(config); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	assertOutputContains(t, filepath.Join(config.OutputDir, "index.html"), `class="non-term"`)
	assertOutputContains(t, filepath.Join(config.OutputDir, "articles", "index.html"), `class="non-term"`)
	assertOutputContains(t, filepath.Join(config.OutputDir, "articles", "first.html"), `class="term"`)
	assertOutputContains(t, filepath.Join(config.OutputDir, "articles", "reference", "syntax.html"), `<h1>Syntax</h1>`)
	if _, err := os.Stat(filepath.Join(config.OutputDir, "about.html")); !os.IsNotExist(err) {
		t.Fatalf("draft output should not exist, stat error = %v", err)
	}
}

// TestBuildTreeUsesIndexPagesAsDirectoryAnchors checks index-based hierarchy and paths.
func TestBuildTreeUsesIndexPagesAsDirectoryAnchors(t *testing.T) {
	root := t.TempDir()
	config := writeSiteFixture(t, root)
	writeFile(t, config.SourceDir, "index.md", page("Home", false, "Home"))
	writeFile(t, config.SourceDir, "articles/index.md", page("Articles", false, "Articles"))
	writeFile(t, config.SourceDir, "articles/first.md", page("First", false, "First"))
	writeFile(t, config.SourceDir, "articles/reference/syntax.md", page("Syntax", false, "Syntax"))

	pages, err := loadPages(config.SourceDir)
	if err != nil {
		t.Fatalf("loadPages() error = %v", err)
	}
	nodes := buildTree(pages)
	byPath := make(map[string]*content.Node)
	for _, node := range nodes {
		byPath[node.SourcePath] = node
		if node.IsPage {
			node.IsTermNode = len(node.Children) == 0
		}
	}

	if got := findPage(t, pages, "index.md").URL; got != "/index.html" {
		t.Errorf("root index URL = %q, want %q", got, "/index.html")
	}
	if got := findPage(t, pages, filepath.Join("articles", "index.md")).OutputPath; got != filepath.Join("articles", "index.html") {
		t.Errorf("section index output = %q", got)
	}
	if len(byPath["index.md"].Children) != 1 {
		t.Errorf("root index children = %d, want 1", len(byPath["index.md"].Children))
	}
	if findPage(t, pages, "index.md").IsTermNode {
		t.Error("root index should be non-terminal when it has child pages")
	}
	if findPage(t, pages, filepath.Join("articles", "first.md")).Parent == nil {
		t.Error("article page should have a parent")
	}
}

// TestBuildReturnsContextForInvalidFrontMatter checks that parsing errors identify their source.
func TestBuildReturnsContextForInvalidFrontMatter(t *testing.T) {
	root := t.TempDir()
	config := writeSiteFixture(t, root)
	writeFile(t, config.SourceDir, "broken.md", "+++\ntitle = [\n+++\nbody")

	err := Build(config)
	if err == nil || !strings.Contains(err.Error(), "broken.md") {
		t.Fatalf("Build() error = %v, want filename context", err)
	}
}

// writeSiteFixture creates isolated source, output, and layout directories for a test.
func writeSiteFixture(t *testing.T, root string) Config {
	t.Helper()
	config := Config{
		SourceDir: filepath.Join(root, "content"),
		OutputDir: filepath.Join(root, "public"),
		LayoutDir: filepath.Join(root, "layout"),
	}
	writeFile(t, config.LayoutDir, "base.html", `{{if .IsTermNode}}{{template "term" .}}{{else}}{{template "non-term" .}}{{end}}`)
	writeFile(t, config.LayoutDir, "term.html", `{{define "term"}}<main class="term">{{.Content}}</main>{{end}}`)
	writeFile(t, config.LayoutDir, "non-term.html", `{{define "non-term"}}<main class="non-term">{{.Content}}</main>{{end}}`)
	return config
}

// writeFile creates parent directories and writes a site fixture file.
func writeFile(t *testing.T, root, relativePath, contents string) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

// page formats Markdown with TOML metadata for use in site tests.
func page(title string, draft bool, body string) string {
	draftValue := "false"
	if draft {
		draftValue = "true"
	}
	return "+++\ntitle = \"" + title + "\"\ndraft = " + draftValue + "\n+++\n" + body
}

// assertOutputContains fails when a generated file omits the expected text.
func assertOutputContains(t *testing.T, path, expected string) {
	t.Helper()
	output, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output %s: %v", path, err)
	}
	if !strings.Contains(string(output), expected) {
		t.Errorf("output %s does not contain %q:\n%s", path, expected, output)
	}
}

// findPage returns the page with the requested source path or fails the test.
func findPage(t *testing.T, pages []*content.Node, sourcePath string) *content.Node {
	t.Helper()
	for _, page := range pages {
		if page.SourcePath == sourcePath {
			return page
		}
	}
	t.Fatalf("page %q not found", sourcePath)
	return nil
}
