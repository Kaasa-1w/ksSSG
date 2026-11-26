package render

import (
	"html/template"
	"strings"
	"testing"
)

func TestMarkdownShortcodesRenderInlineAndBlockAndSkipCode(t *testing.T) {
	source := []byte("before\n" +
		`{{< tag text="two words" >}}` + "\n" +
		`after {{< tag text=inline >}}` + "\n\n" +
		"`{{< tag text=inline-code >}}`\n\n" +
		"```html\n{{< tag text=fenced >}}\n```\n\n" +
		"    {{< tag text=indented >}}\n")
	var calls []string
	got, err := markdownToHTML(source, func(name string, args map[string]string) (template.HTML, error) {
		if name != "tag" {
			t.Fatalf("shortcode name = %q, want tag", name)
		}
		calls = append(calls, args["text"])
		return template.HTML("<b>" + args["text"] + "</b>"), nil
	})
	if err != nil {
		t.Fatalf("markdownToHTML() error = %v", err)
	}

	if strings.Contains(string(got), "<p>before\n<b>two words</b>") {
		t.Fatalf("block shortcode was wrapped in the preceding paragraph:\n%s", got)
	}
	for _, expected := range []string{
		"<p>before</p>\n<b>two words</b>\n",
		"<p>after <b>inline</b></p>\n",
		"<code>{{&lt; tag text=inline-code &gt;}}</code>",
		"<pre><code class=\"language-html\">{{&lt; tag text=fenced &gt;}}\n</code></pre>",
		"<pre><code>{{&lt; tag text=indented &gt;}}\n\n</code></pre>",
	} {
		if !strings.Contains(string(got), expected) {
			t.Errorf("rendered Markdown does not contain %q:\n%s", expected, got)
		}
	}
	if len(calls) != 2 || calls[0] != "two words" || calls[1] != "inline" {
		t.Errorf("shortcode calls = %v, want [two words inline]", calls)
	}
}

func TestMarkdownShortcodesReportMalformedCalls(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "missing argument value", source: `{{< badge label= >}}`},
		{name: "unclosed quote", source: `{{< badge label="hello >}}`},
		{name: "duplicate argument", source: `{{< badge label=one label=two >}}`},
		{name: "missing closing marker", source: `{{< badge label=one`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := markdownToHTML([]byte(test.source), func(string, map[string]string) (template.HTML, error) {
				t.Fatal("shortcode executor should not run for malformed calls")
				return "", nil
			})
			if err == nil {
				t.Fatal("markdownToHTML() error = nil, want malformed shortcode error")
			}
		})
	}
}

func TestMarkdownToHTMLRejectsShortcodesWithoutRenderer(t *testing.T) {
	_, err := MarkdownToHTML([]byte(`{{< badge >}}`))
	if err == nil || !strings.Contains(err.Error(), "without a shortcode renderer") {
		t.Fatalf("MarkdownToHTML() error = %v, want missing shortcode renderer error", err)
	}
}
