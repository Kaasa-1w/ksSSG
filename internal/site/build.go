package site

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ks-ssg/internal/content"
	"ks-ssg/internal/render"
)

type Config struct {
	SourceDir string
	OutputDir string
	LayoutDir string
}

// DefaultConfig returns the repository's default source, output, and layout directories.
func DefaultConfig() Config {
	return Config{
		SourceDir: "test/content",
		OutputDir: "test/public",
		LayoutDir: "test/layout",
	}
}

// Build reads the site's pages, creates their hierarchy, and renders published content.
func Build(config Config) error {
	if err := config.validate(); err != nil {
		return err
	}

	renderer, err := render.NewRenderer(config.LayoutDir)
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	pages, err := loadPages(config.SourceDir)
	if err != nil {
		return err
	}

	nodes := buildTree(pages)
	for _, node := range nodes {
		node.IsTermNode = len(node.Children) == 0
	}

	for _, page := range pages {
		if page.Draft {
			continue
		}

		page.Content, err = render.MarkdownToHTML(page.Body)
		if err != nil {
			return fmt.Errorf("failed to convert Markdown for %q: %w", page.SourcePath, err)
		}

		outputPath := filepath.Join(config.OutputDir, page.OutputPath)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory for %q: %w", outputPath, err)
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", outputPath, err)
		}

		renderErr := renderer.Render(file, page)
		closeErr := file.Close()
		if renderErr != nil {
			return fmt.Errorf("failed to render %q: %w", outputPath, renderErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close output file %q: %w", outputPath, closeErr)
		}
	}
	return nil
}

// validate ensures that all required site directories are configured.
func (config Config) validate() error {
	if config.SourceDir == "" {
		return fmt.Errorf("source directory is not configured")
	}
	if config.OutputDir == "" {
		return fmt.Errorf("output directory is not configured")
	}
	if config.LayoutDir == "" {
		return fmt.Errorf("layout directory is not configured")
	}
	return nil
}

// loadPages reads Markdown files and derives each page's output path and URL.
func loadPages(sourceDir string) ([]*content.Node, error) {
	var pages []*content.Node
	err := filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to walk %q: %w", path, walkErr)
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %q: %w", path, err)
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %q: %w", path, err)
		}

		frontMatter, body, err := content.ParseFrontMatter(source)
		if err != nil {
			return fmt.Errorf("failed to parse front matter in %q: %w", path, err)
		}

		outputPath := strings.TrimSuffix(relativePath, filepath.Ext(relativePath)) + ".html"
		pages = append(pages, &content.Node{
			IsPage:     true,
			SourcePath: relativePath,
			OutputPath: outputPath,
			URL:        "/" + filepath.ToSlash(outputPath),
			Title:      frontMatter.Title,
			Date:       frontMatter.Date,
			Draft:      frontMatter.Draft,
			Body:       body,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory %q: %w", sourceDir, err)
	}
	return pages, nil
}

// buildTree links pages and directory nodes into a hierarchy.
func buildTree(pages []*content.Node) []*content.Node {
	directories := make(map[string]struct{})
	indexPages := make(map[string]*content.Node)
	for _, page := range pages {
		directory := filepath.Dir(page.SourcePath)
		if filepath.Base(page.SourcePath) == "index.md" {
			indexPages[directory] = page
		}
		for directory != "." {
			directories[directory] = struct{}{}
			directory = filepath.Dir(directory)
		}
	}

	directoryPaths := make([]string, 0, len(directories))
	for directory := range directories {
		directoryPaths = append(directoryPaths, directory)
	}
	sort.Strings(directoryPaths)

	anchors := make(map[string]*content.Node, len(directoryPaths)+1)
	for directory, page := range indexPages {
		anchors[directory] = page
	}
	nodes := append([]*content.Node(nil), pages...)
	for _, directory := range directoryPaths {
		if _, ok := anchors[directory]; ok {
			continue
		}
		node := &content.Node{SourcePath: directory}
		anchors[directory] = node
		nodes = append(nodes, node)
	}

	for _, directory := range directoryPaths {
		node := anchors[directory]
		parent := anchors[filepath.Dir(directory)]
		attach(parent, node)
	}
	for _, page := range pages {
		if filepath.Base(page.SourcePath) == "index.md" {
			continue
		}
		attach(anchors[filepath.Dir(page.SourcePath)], page)
	}
	return nodes
}

// attach adds a child node to its parent when both nodes are valid.
func attach(parent, child *content.Node) {
	if parent == nil || child == nil || parent == child {
		return
	}
	child.Parent = parent
	parent.Children = append(parent.Children, child)
}
