package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBuildsUsingCommandLineDirectories(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "content")
	outputDir := filepath.Join(root, "public")
	layoutDir := filepath.Join(root, "layout")

	writeFile(t, sourceDir, "index.md", "+++\ntitle = \"Home\"\n+++\n# Home")
	writeFile(t, layoutDir, "base.html", `{{if .IsTermNode}}{{template "term" .}}{{else}}{{template "non-term" .}}{{end}}`)
	writeFile(t, layoutDir, "term.html", `{{define "term"}}{{.Content}}{{end}}`)
	writeFile(t, layoutDir, "non-term.html", `{{define "non-term"}}{{.Content}}{{end}}`)

	if err := run([]string{
		"build",
		"-source-dir", sourceDir,
		"-output-dir", outputDir,
		"-layout-dir", layoutDir,
	}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("read generated page: %v", err)
	}
	if got, want := string(output), "<h1>Home</h1>\n"; got != want {
		t.Errorf("generated page = %q, want %q", got, want)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	if err := run([]string{"publish"}); err == nil {
		t.Fatal("run() error = nil, want unknown command error")
	}
}

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
