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
}

func initializeProject(directory string) error {
	path := filepath.Join(directory, configFileName)
	config := projectConfig{
		SourceDir: "content",
		OutputDir: "public",
		LayoutDir: "layout",
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
	configPath := filepath.Join(directory, configFileName)
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read configuration file %q: %w", configPath, err)
	}

	var project projectConfig
	if err := json.Unmarshal(contents, &project); err != nil {
		return fmt.Errorf("parse configuration file %q: %w", configPath, err)
	}
	if project.SourceDir == "" {
		return fmt.Errorf("configuration field %q must not be empty", "source_dir")
	}
	if project.OutputDir == "" {
		return fmt.Errorf("configuration field %q must not be empty", "output_dir")
	}
	if project.LayoutDir == "" {
		return fmt.Errorf("configuration field %q must not be empty", "layout_dir")
	}

	baseDir, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("resolve configuration directory %q: %w", directory, err)
	}
	config := site.Config{
		SourceDir: resolveConfigPath(baseDir, project.SourceDir),
		OutputDir: resolveConfigPath(baseDir, project.OutputDir),
		LayoutDir: resolveConfigPath(baseDir, project.LayoutDir),
	}
	return site.Build(config)
}

func resolveConfigPath(baseDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}
