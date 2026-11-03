package render

import "testing"

// TestMarkdownToHTML verifies the generated HTML for a Markdown heading.
func TestMarkdownToHTML(t *testing.T) {
	got, err := MarkdownToHTML([]byte("# Title"))
	if err != nil {
		t.Fatalf("MarkdownToHTML() error = %v", err)
	}
	if want := "<h1>Title</h1>\n"; string(got) != want {
		t.Errorf("MarkdownToHTML() = %q, want %q", got, want)
	}
}
