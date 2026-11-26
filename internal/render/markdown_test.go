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

func TestMarkdownToHTMLJapanese(t *testing.T) {
	got, err := MarkdownToHTML([]byte("# サンプル\n\nこれは*日本語*です。"))
	if err != nil {
		t.Fatalf("MarkdownToHTML() error = %v", err)
	}
	want := "<h1>サンプル</h1>\n<p>これは<em>日本語</em>です。</p>\n"
	if string(got) != want {
		t.Errorf("MarkdownToHTML() = %q, want %q", got, want)
	}
}

func TestMarkdownToHTMLBlocksAndInline(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "paragraph and emphasis",
			source: "Hello *world* and **friends**",
			want:   "<p>Hello <em>world</em> and <strong>friends</strong></p>\n",
		},
		{
			name:   "unordered list",
			source: "- one\n- two",
			want:   "<ul>\n<li>one</li>\n<li>two</li>\n</ul>\n",
		},
		{
			name:   "ordered list",
			source: "1. first\n2. second",
			want:   "<ol>\n<li>first</li>\n<li>second</li>\n</ol>\n",
		},
		{
			name:   "setext heading",
			source: "Title\n===",
			want:   "<h1>Title</h1>\n",
		},
		{
			name:   "blockquote",
			source: "> quoted\n> text",
			want:   "<blockquote>\n<p>quoted\ntext</p>\n</blockquote>\n",
		},
		{
			name:   "fenced code",
			source: "```go\nfmt.Println(\"hello\")\n```",
			want:   "<pre><code class=\"language-go\">fmt.Println(\"hello\")\n</code></pre>\n",
		},
		{
			name:   "link",
			source: "[site](https://example.com)",
			want:   `<p><a href="https://example.com">site</a></p>` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkdownToHTML([]byte(tt.source))
			if err != nil {
				t.Fatalf("MarkdownToHTML() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarkdownToHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}
