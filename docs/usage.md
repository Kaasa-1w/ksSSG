# Using ks-ssg

## Initialize a project

Build the CLI from the repository source:

```sh
go build .
```

This creates `ks-ssg` (`ks-ssg.exe` on Windows). Run it from the directory that should contain the site:

```sh
ks-ssg init
ks-ssg build
```

`init` creates `ks-ssg.json` and any configured directories that do not exist yet (`content/`, `public/`, `layout/`, and `static/` by default). Existing directories and their contents are left alone. Add your Markdown pages and layout templates before running `build`; `build` reads the configuration from the current directory and writes the generated site to the configured output directory.

## Configure the site

The generated `ks-ssg.json` looks like this:

```json
{
  "source_dir": "content",
  "output_dir": "public",
  "layout_dir": "layout",
  "static_dir": "static",
  "base_url": "https://example.com/"
}
```

Paths are relative to the directory containing `ks-ssg.json`; absolute paths are also accepted.

| Key | Purpose |
| --- | --- |
| `source_dir` | Markdown files to build. Required. |
| `output_dir` | Destination for generated HTML and copied static files. Required. |
| `layout_dir` | HTML templates used to render pages. Required. |
| `static_dir` | Files copied into the output directory, preserving subdirectories. Defaults to `static` if omitted. If the directory does not exist, copying is skipped. |
| `base_url` | The site's root URL. `init` uses the placeholder `https://example.com/`; change it to the deployed site's URL. Older configurations may omit it. |

Page templates can refer to the configured root URL as `.BaseURL`. It is passed to shortcode templates as `.Page.BaseURL` as well; see [Markdown shortcodes](shortcodes.md).

## Add content

Put `.md` files under the configured source directory. Each file starts with TOML front matter between `+++` lines:

```toml
+++
title = "About"
date = "2026/10/30"
draft = false
+++
```

Write Markdown after the closing delimiter. The supported front matter fields are `title` and `date` (strings), and `draft` (boolean). Draft pages are not emitted. A source file such as `articles/first.md` becomes `public/articles/first.html`. An `index.md` represents its directory in the page hierarchy.

The Markdown renderer supports headings, paragraphs, emphasis, links, images, lists, blockquotes, thematic breaks, fenced and indented code blocks, and raw HTML. User-defined shortcodes are documented separately.

## Add layouts

The layout directory must contain `base.html` and templates named `term` and `non-term` (normally defined in `term.html` and `non-term.html`). All `.html` files under the layout directory, including nested directories, are loaded together, so templates can call definitions from other files.

`base.html` provides the shared document wrapper and renders `.Content`:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{.Title}}</title>
</head>
<body>
  <a href="{{.BaseURL}}">Home</a>
  {{.Content}}
</body>
</html>
```

The program chooses `term` for a page with no child pages and `non-term` for a page with children. Each template renders the page body inside the shared wrapper. For example:

```html
{{define "term"}}<main class="term">{{.Content}}</main>{{end}}
```

```html
{{define "non-term"}}<main class="non-term">{{.Content}}</main>{{end}}
```

Page data available to layout templates includes `.Title`, `.Date`, `.URL`, `.BaseURL`, `.Content`, and `.Children`.

## Add static files

Place CSS, images, and other assets under `static/`. During a build, each regular file is copied to the matching path under `public/`; for example, `static/css/site.css` becomes `public/css/site.css`.
