package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunBuildsUsingProjectConfiguration(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "content")
	outputDir := filepath.Join(root, "public")
	layoutDir := filepath.Join(root, "layout")

	if err := runAt([]string{"init"}, root); err != nil {
		t.Fatalf("runAt(init) error = %v", err)
	}
	writeFile(t, sourceDir, "index.md", "+++\ntitle = \"Home\"\n+++\n# Home")
	writeFile(t, layoutDir, "base.html", `{{if .IsTermNode}}{{template "term" .}}{{else}}{{template "non-term" .}}{{end}}`)
	writeFile(t, layoutDir, "term.html", `{{define "term"}}{{.Content}}{{end}}`)
	writeFile(t, layoutDir, "non-term.html", `{{define "non-term"}}{{.Content}}{{end}}`)

	if err := runAt([]string{"build"}, root); err != nil {
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

func TestInitDoesNotOverwriteExistingConfiguration(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, configFileName)
	writeFile(t, root, configFileName, `{"source_dir":"custom"}`)

	if err := runAt([]string{"init"}, root); err == nil {
		t.Fatal("runAt(init) error = nil, want existing configuration error")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read existing configuration: %v", err)
	}
	if got, want := string(contents), `{"source_dir":"custom"}`; got != want {
		t.Errorf("configuration = %q, want unchanged %q", got, want)
	}
}

func TestInitCreatesConfigurationWithDefaultDirectories(t *testing.T) {
	root := t.TempDir()
	if err := runAt([]string{"init"}, root); err != nil {
		t.Fatalf("runAt(init) error = %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(root, configFileName))
	if err != nil {
		t.Fatalf("read generated configuration: %v", err)
	}
	var config projectConfig
	if err := json.Unmarshal(contents, &config); err != nil {
		t.Fatalf("parse generated configuration: %v", err)
	}
	if config != (projectConfig{SourceDir: "content", OutputDir: "public", LayoutDir: "layout"}) {
		t.Errorf("generated configuration = %+v", config)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	if err := runAt([]string{"publish"}, t.TempDir()); err == nil {
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
