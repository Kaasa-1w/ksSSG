package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ks-ssg/internal/site"
)

const configFileName = "ks-ssg.json"

type projectConfig struct {
	SourceDir string `json:"source_dir"`
	OutputDir string `json:"output_dir"`
	LayoutDir string `json:"layout_dir"`
	StaticDir string `json:"static_dir"`
	BaseURL   string `json:"base_url"`
}

func initializeProject(directory string) error {
	path := filepath.Join(directory, configFileName)
	config := projectConfig{
		SourceDir: "content",
		OutputDir: "public",
		LayoutDir: "layout",
		StaticDir: "static",
		BaseURL:   "https://example.com/",
	}
	contents, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode project configuration: %w", err)
	}
	contents = append(contents, '\n')

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("create configuration file %q: %w", path, err)
	}
	_, writeErr := file.Write(contents)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write configuration file %q: %w", path, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close configuration file %q: %w", path, closeErr)
	}

	fmt.Printf("Created configuration file: %s\n", path)
	return nil
}

func buildProject(directory string) error {
	config, err := loadProjectConfig(directory)
	if err != nil {
		return err
	}
	return site.Build(config)
}

func loadProjectConfig(directory string) (site.Config, error) {
	configPath := filepath.Join(directory, configFileName)
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return site.Config{}, fmt.Errorf("read configuration file %q: %w", configPath, err)
	}

	var project projectConfig
	if err := json.Unmarshal(contents, &project); err != nil {
		return site.Config{}, fmt.Errorf("parse configuration file %q: %w", configPath, err)
	}
	if err := project.validate(); err != nil {
		return site.Config{}, err
	}

	baseDir, err := filepath.Abs(directory)
	if err != nil {
		return site.Config{}, fmt.Errorf("resolve configuration directory %q: %w", directory, err)
	}
	return project.toSiteConfig(baseDir), nil
}

func (config projectConfig) validate() error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "source_dir", value: config.SourceDir},
		{name: "output_dir", value: config.OutputDir},
		{name: "layout_dir", value: config.LayoutDir},
	}
	for _, field := range fields {
		if field.value == "" {
			return fmt.Errorf("configuration field %q must not be empty", field.name)
		}
	}
	return nil
}

func (config projectConfig) toSiteConfig(baseDir string) site.Config {
	staticDir := config.StaticDir
	if staticDir == "" {
		staticDir = "static"
	}
	return site.Config{
		SourceDir: resolveConfigPath(baseDir, config.SourceDir),
		OutputDir: resolveConfigPath(baseDir, config.OutputDir),
		LayoutDir: resolveConfigPath(baseDir, config.LayoutDir),
		StaticDir: resolveConfigPath(baseDir, staticDir),
		BaseURL:   config.BaseURL,
	}
}

func resolveConfigPath(baseDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}
