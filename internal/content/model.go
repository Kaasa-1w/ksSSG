package content

import "html/template"

type Node struct {
	IsTermNode bool
	IsPage     bool

	SourcePath string
	OutputPath string
	URL        string

	Title   string
	Date    string
	Content template.HTML
	Body    []byte
	Draft   bool

	Parent   *Node
	Children []*Node
}
