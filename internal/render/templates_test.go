package render

import (
	"bytes"
	"html/template"
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
