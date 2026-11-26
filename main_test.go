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

func TestLoadProjectConfigRejectsEmptyDirectories(t *testing.T) {
	tests := []struct {
		name   string
		config string
		field  string
	}{
		{name: "source", config: `{"source_dir":"","output_dir":"public","layout_dir":"layout"}`, field: "source_dir"},
		{name: "output", config: `{"source_dir":"content","output_dir":"","layout_dir":"layout"}`, field: "output_dir"},
		{name: "layout", config: `{"source_dir":"content","output_dir":"public","layout_dir":""}`, field: "layout_dir"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, configFileName, test.config)

			_, err := loadProjectConfig(root)
			if err == nil || err.Error() != `configuration field "`+test.field+`" must not be empty` {
				t.Fatalf("loadProjectConfig() error = %v, want empty %s error", err, test.field)
			}
		})
	}
}

func TestLoadProjectConfigResolvesRelativeAndAbsoluteDirectories(t *testing.T) {
	root := t.TempDir()
	absoluteSourceDir := filepath.Join(root, "external-content")
	contents, err := json.Marshal(projectConfig{
		SourceDir: absoluteSourceDir,
		OutputDir: "public",
		LayoutDir: "layout",
	})
	if err != nil {
		t.Fatalf("marshal project configuration: %v", err)
	}
	writeFile(t, root, configFileName, string(contents))

	config, err := loadProjectConfig(root)
	if err != nil {
		t.Fatalf("loadProjectConfig() error = %v", err)
	}
	if config.SourceDir != absoluteSourceDir {
		t.Errorf("source directory = %q, want absolute path %q", config.SourceDir, absoluteSourceDir)
	}
	if config.OutputDir != filepath.Join(root, "public") {
		t.Errorf("output directory = %q, want %q", config.OutputDir, filepath.Join(root, "public"))
	}
	if config.LayoutDir != filepath.Join(root, "layout") {
		t.Errorf("layout directory = %q, want %q", config.LayoutDir, filepath.Join(root, "layout"))
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
