# Markdown shortcodes

Markdown pages can call named templates from `layout/` using shortcodes. Define a shortcode in any `.html` file under `layout/`:

The site's root URL is configured as `base_url` in `ks-ssg.json` (initialized to `https://example.com/`). Page templates can use it as `.BaseURL`; shortcode templates can use `.Page.BaseURL`.

```html
{{define "shortcode:link"}}
{{if eq .Args.route "about"}}
<a href="/about.html">
{{else}}
<a href="{{.Args.url}}">
{{end}}
  {{if .Args.label}}{{.Args.label}}{{else}}{{.Page.Title}}{{end}}
</a>
{{end}}
```

Call it from a Markdown page with named string arguments:

```md
Read about {{< link route="about" label="About" >}}.
```

A shortcode can also appear on its own line. Its output is emitted as a block instead of being wrapped in a paragraph:

```md
{{< link url="/about.html" label="About" >}}
```

The template receives:

- `.Page`: the current page, including fields such as `Title`, `URL`, `Date`, and `BaseURL`.
- `.Args`: a map of string arguments supplied in the shortcode call. Values may be quoted to include spaces; unquoted values end at whitespace.

Shortcodes are regular `html/template` templates, so conditionals and nested template calls are available. For example, a template can conditionally choose a URL or HTML attributes with `if` and `else`. Use `.Args.name` and `.Page` values as template data; interpolated values are contextually escaped by `html/template`.

Shortcodes are not expanded inside inline code, fenced code blocks, or indented code blocks. Calls to undefined shortcodes, malformed syntax, and template execution errors fail the build and identify the Markdown source file.
