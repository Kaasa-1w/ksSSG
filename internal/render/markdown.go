package render

import (
	"html/template"
	"strings"
	"unicode/utf8"
)

type shortcodeExecutor func(string, map[string]string) (template.HTML, error)

type markdownRenderState struct {
	executeShortcode shortcodeExecutor
	err              error
}

// MarkdownToHTML converts Markdown source into trusted HTML.
func MarkdownToHTML(source []byte) (template.HTML, error) {
	return markdownToHTML(source, nil)
}

func markdownToHTML(source []byte, executeShortcode shortcodeExecutor) (template.HTML, error) {
	text := normalizeNewlines(string(source))

	var out strings.Builder
	state := &markdownRenderState{executeShortcode: executeShortcode}
	renderBlocks(strings.Split(text, "\n"), &out, state)
	if state.err != nil {
		return "", state.err
	}
	return template.HTML(out.String()), nil
}

// normalizeNewlines converts all supported line endings to LF.
func normalizeNewlines(source string) string {
	return strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n")
}

// renderBlocks parses and writes block-level Markdown elements.
func renderBlocks(lines []string, out *strings.Builder, state *markdownRenderState) {
	for i := 0; i < len(lines); {
		if state.err != nil {
			return
		}
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		trimmedLeft := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmedLeft, "```") || strings.HasPrefix(trimmedLeft, "~~~") {
			i = renderFencedCode(lines, i, out)
			continue
		}

		trimmed := strings.TrimSpace(line)
		if isThematicBreak(trimmed) {
			out.WriteString("<hr>\n")
			i++
			continue
		}

		if heading, ok := parseATXHeading(line); ok {
			writeHeading(out, heading.level, heading.text, state)
			i++
			continue
		}

		if level, ok := parseSetextHeading(lines, i); ok {
			writeHeading(out, level, strings.TrimSpace(line), state)
			i += 2
			continue
		}

		if strings.HasPrefix(line, ">") || strings.HasPrefix(line, " >") {
			i = renderBlockquote(lines, i, out, state)
			continue
		}
		if _, ok := unorderedMarker(line); ok {
			i = renderUnorderedList(lines, i, out, state)
			continue
		}
		if _, ok := orderedMarker(line); ok {
			i = renderOrderedList(lines, i, out, state)
			continue
		}
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			i = renderIndentedCode(lines, i, out)
			continue
		}

		call, isShortcodeBlock, err := parseStandaloneShortcode(line)
		if err != nil {
			state.err = err
			return
		}
		if isShortcodeBlock {
			output := state.renderShortcode(call)
			out.WriteString(output)
			if state.err != nil {
				return
			}
			if !strings.HasSuffix(output, "\n") {
				out.WriteByte('\n')
			}
			i++
			continue
		}

		i = renderParagraph(lines, i, out, state)
	}
}

type heading struct {
	level int
	text  string
}

// writeHeading writes a heading and renders inline Markdown in its text.
func writeHeading(out *strings.Builder, level int, text string, state *markdownRenderState) {
	out.WriteString("<h")
	out.WriteByte('0' + byte(level))
	out.WriteByte('>')
	out.WriteString(renderInline(text, state))
	out.WriteString("</h")
	out.WriteByte('0' + byte(level))
	out.WriteString(">\n")
}

// parseATXHeading recognizes a hash-prefixed ATX heading.
func parseATXHeading(line string) (heading, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) == 0 || trimmed[0] != '#' {
		return heading{}, false
	}
	level := 0
	for level < len(trimmed) && level < 6 && trimmed[level] == '#' {
		level++
	}
	if level == 0 || (level < len(trimmed) && trimmed[level] != ' ' && trimmed[level] != '\t') {
		return heading{}, false
	}
	text := strings.TrimSpace(trimmed[level:])
	if strings.HasSuffix(text, "#") {
		text = strings.TrimSpace(strings.TrimSuffix(text, "#"))
	}
	return heading{level: level, text: text}, true
}

// parseSetextHeading identifies a Setext underline following the current line.
func parseSetextHeading(lines []string, start int) (int, bool) {
	if start+1 >= len(lines) {
		return 0, false
	}
	underline := strings.TrimSpace(lines[start+1])
	if isSetextUnderline(underline, '=') {
		return 1, true
	}
	if isSetextUnderline(underline, '-') {
		return 2, true
	}
	return 0, false
}

// isSetextUnderline reports whether a line contains only one Setext underline marker.
func isSetextUnderline(line string, marker byte) bool {
	if line == "" {
		return false
	}
	found := false
	for _, r := range line {
		if r == rune(marker) {
			found = true
		} else if r != ' ' && r != '\t' {
			return false
		}
	}
	return found
}

// isThematicBreak reports whether a line is a Markdown thematic break.
func isThematicBreak(line string) bool {
	if line == "" {
		return false
	}
	marker := rune(0)
	count := 0
	for _, r := range line {
		if r == ' ' || r == '\t' {
			continue
		}
		if r != '*' && r != '-' && r != '_' {
			return false
		}
		if marker == 0 {
			marker = r
		} else if r != marker {
			return false
		}
		count++
	}
	return count >= 3
}

// renderFencedCode collects a fenced code block and advances past its closing fence.
func renderFencedCode(lines []string, start int, out *strings.Builder) int {
	trimmed := strings.TrimLeft(lines[start], " ")
	fence := trimmed[:3]
	info := strings.TrimSpace(trimmed[3:])
	i := start + 1
	var code strings.Builder
	for i < len(lines) {
		if strings.HasPrefix(strings.TrimLeft(lines[i], " "), fence) {
			i++
			break
		}
		code.WriteString(lines[i])
		code.WriteByte('\n')
		i++
	}

	language := ""
	if info != "" {
		language = strings.Fields(info)[0]
	}
	writeCodeBlock(out, language, code.String())
	return i
}

// writeCodeBlock writes escaped code using an optional language class.
func writeCodeBlock(out *strings.Builder, language, code string) {
	out.WriteString("<pre><code")
	if language != "" {
		out.WriteString(` class="language-`)
		out.WriteString(escapeHTMLAttribute(language))
		out.WriteByte('"')
	}
	out.WriteByte('>')
	out.WriteString(escapeHTML(code))
	out.WriteString("</code></pre>\n")
}

// renderIndentedCode collects an indented code block and advances to the next block.
func renderIndentedCode(lines []string, start int, out *strings.Builder) int {
	var code strings.Builder
	i := start
	for i < len(lines) {
		line := lines[i]
		switch {
		case strings.HasPrefix(line, "    "):
			code.WriteString(line[4:])
			code.WriteByte('\n')
		case strings.HasPrefix(line, "\t"):
			code.WriteString(line[1:])
			code.WriteByte('\n')
		case strings.TrimSpace(line) == "":
			code.WriteByte('\n')
		default:
			writeCodeBlock(out, "", code.String())
			return i
		}
		i++
	}
	writeCodeBlock(out, "", code.String())
	return i
}

// renderBlockquote collects quoted lines and renders them as nested blocks.
func renderBlockquote(lines []string, start int, out *strings.Builder, state *markdownRenderState) int {
	var quoted []string
	i := start
	for i < len(lines) {
		line := lines[i]
		if strings.HasPrefix(line, ">") {
			quoted = append(quoted, strings.TrimPrefix(strings.TrimPrefix(line, ">"), " "))
		} else if strings.HasPrefix(line, " >") {
			quoted = append(quoted, strings.TrimPrefix(strings.TrimPrefix(line, " "), "> "))
		} else {
			break
		}
		i++
	}
	out.WriteString("<blockquote>\n")
	renderBlocks(quoted, out, state)
	out.WriteString("</blockquote>\n")
	return i
}

// unorderedMarker reports whether a line starts an unordered-list item.
func unorderedMarker(line string) (string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) >= 2 && (trimmed[0] == '-' || trimmed[0] == '*' || trimmed[0] == '+') &&
		(trimmed[1] == ' ' || trimmed[1] == '\t') {
		return string(trimmed[0]), true
	}
	return "", false
}

// orderedMarker reports whether a line starts an ordered-list item.
func orderedMarker(line string) (string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	n := 0
	for n < len(trimmed) && n < 9 && trimmed[n] >= '0' && trimmed[n] <= '9' {
		n++
	}
	if n > 0 && n+1 < len(trimmed) && (trimmed[n] == '.' || trimmed[n] == ')') &&
		(trimmed[n+1] == ' ' || trimmed[n+1] == '\t') {
		return trimmed[:n], true
	}
	return "", false
}

// renderUnorderedList renders consecutive unordered-list items.
func renderUnorderedList(lines []string, start int, out *strings.Builder, state *markdownRenderState) int {
	out.WriteString("<ul>\n")
	i := start
	for i < len(lines) {
		if _, ok := unorderedMarker(lines[i]); !ok {
			break
		}
		trimmed := strings.TrimLeft(lines[i], " ")
		text := strings.TrimLeft(trimmed[1:], " \t")
		out.WriteString("<li>")
		out.WriteString(renderInline(text, state))
		out.WriteString("</li>\n")
		i++
	}
	out.WriteString("</ul>\n")
	return i
}

// renderOrderedList renders consecutive ordered-list items.
func renderOrderedList(lines []string, start int, out *strings.Builder, state *markdownRenderState) int {
	out.WriteString("<ol>\n")
	i := start
	for i < len(lines) {
		marker, ok := orderedMarker(lines[i])
		if !ok {
			break
		}
		trimmed := strings.TrimLeft(lines[i], " ")
		text := strings.TrimSpace(trimmed[len(marker):])
		text = strings.TrimLeft(text, ". ")
		out.WriteString("<li>")
		out.WriteString(renderInline(text, state))
		out.WriteString("</li>\n")
		i++
	}
	out.WriteString("</ol>\n")
	return i
}

// renderParagraph collects paragraph lines and renders inline content with hard breaks.
func renderParagraph(lines []string, start int, out *strings.Builder, state *markdownRenderState) int {
	var paragraph []string
	i := start
	for i < len(lines) {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			break
		}
		if i > start && beginsBlock(line) {
			break
		}
		paragraph = append(paragraph, line)
		i++
	}
	out.WriteString("<p>")
	for n, line := range paragraph {
		if n > 0 {
			out.WriteByte('\n')
		}
		hardBreak := strings.HasSuffix(line, "  ")
		if hardBreak {
			line = strings.TrimSuffix(line, "  ")
			out.WriteString(renderInline(line, state))
			out.WriteString("<br>")
		} else {
			out.WriteString(renderInline(line, state))
		}
	}
	out.WriteString("</p>\n")
	return i
}

// beginsBlock reports whether a line starts a block-level Markdown element.
func beginsBlock(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || isThematicBreak(trimmed) {
		return true
	}
	if strings.HasPrefix(trimmed, "{{<") {
		_, isShortcodeBlock, err := parseStandaloneShortcode(line)
		if isShortcodeBlock || err != nil {
			return true
		}
	}
	if _, ok := parseATXHeading(line); ok {
		return true
	}
	left := strings.TrimLeft(line, " ")
	if strings.HasPrefix(left, "```") || strings.HasPrefix(left, "~~~") {
		return true
	}
	if _, ok := unorderedMarker(line); ok {
		return true
	}
	if _, ok := orderedMarker(line); ok {
		return true
	}
	return strings.HasPrefix(line, ">") || strings.HasPrefix(line, " >") ||
		strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t")
}

// renderInline parses emphasis, links, code, escapes, and raw HTML within text.
func renderInline(source string, state *markdownRenderState) string {
	var out strings.Builder
	for i := 0; i < len(source); {
		if state.err != nil {
			return out.String()
		}
		if source[i] == '\\' && i+1 < len(source) && isEscapable(source[i+1]) {
			out.WriteRune(rune(source[i+1]))
			i += 2
			continue
		}

		if source[i] == '`' {
			end := strings.IndexByte(source[i+1:], '`')
			if end >= 0 {
				end += i + 1
				code := source[i+1 : end]
				out.WriteString("<code>")
				out.WriteString(escapeHTML(code))
				out.WriteString("</code>")
				i = end + 1
				continue
			}
		}

		if strings.HasPrefix(source[i:], "{{<") {
			call, end, err := parseShortcodeAt(source, i)
			if err != nil {
				state.err = err
				return out.String()
			}
			out.WriteString(state.renderShortcode(call))
			if state.err != nil {
				return out.String()
			}
			i = end
			continue
		}

		if source[i] == '!' && i+1 < len(source) && source[i+1] == '[' {
			if alt, destination, end, ok := parseLink(source, i+1); ok {
				out.WriteString(`<img src="`)
				out.WriteString(escapeHTMLAttribute(destination))
				out.WriteString(`" alt="`)
				out.WriteString(escapeHTMLAttribute(alt))
				out.WriteString(`">`)
				i = end
				continue
			}
		}

		if source[i] == '[' {
			if text, destination, end, ok := parseLink(source, i); ok {
				out.WriteString(`<a href="`)
				out.WriteString(escapeHTMLAttribute(destination))
				out.WriteString(`">`)
				out.WriteString(renderInline(text, state))
				out.WriteString("</a>")
				i = end
				continue
			}
		}

		if i+1 < len(source) && (strings.HasPrefix(source[i:], "**") || strings.HasPrefix(source[i:], "__")) {
			marker := source[i : i+2]
			if end := strings.Index(source[i+2:], marker); end >= 0 {
				end += i + 2
				out.WriteString("<strong>")
				out.WriteString(renderInline(source[i+2:end], state))
				out.WriteString("</strong>")
				i = end + 2
				continue
			}
		}

		if source[i] == '*' || source[i] == '_' {
			marker := source[i]
			if end := strings.IndexByte(source[i+1:], marker); end >= 0 {
				end += i + 1
				out.WriteString("<em>")
				out.WriteString(renderInline(source[i+1:end], state))
				out.WriteString("</em>")
				i = end + 1
				continue
			}
		}

		if strings.HasPrefix(source[i:], "<http://") || strings.HasPrefix(source[i:], "<https://") ||
			strings.HasPrefix(source[i:], "<mailto:") {
			if end := strings.IndexByte(source[i+1:], '>'); end >= 0 {
				end += i + 1
				url := source[i+1 : end]
				out.WriteString(`<a href="`)
				out.WriteString(escapeHTMLAttribute(url))
				out.WriteString(`">`)
				out.WriteString(escapeHTML(url))
				out.WriteString("</a>")
				i = end + 1
				continue
			}
		}

		if source[i] == '<' && isRawHTMLAt(source, i) {
			end := findRawHTMLEnd(source, i)
			if end > i {
				out.WriteString(source[i:end])
				i = end
				continue
			}
		}

		if strings.HasPrefix(source[i:], "  \n") {
			out.WriteString("<br>\n")
			i += 3
			continue
		}

		r, size := utf8.DecodeRuneInString(source[i:])
		out.WriteString(escapeHTML(string(r)))
		i += size
	}
	return out.String()
}

// parseLink extracts link text and destination starting at a bracket or image marker.
func parseLink(source string, start int) (text string, destination string, end int, ok bool) {
	open := start
	if source[open] == '!' {
		open++
	}
	closeBracket := findMatching(source, open, '[', ']')
	if closeBracket < 0 || closeBracket+1 >= len(source) || source[closeBracket+1] != '(' {
		return "", "", 0, false
	}
	closeParen := findMatching(source, closeBracket+1, '(', ')')
	if closeParen < 0 {
		return "", "", 0, false
	}
	text = source[open+1 : closeBracket]
	destination = strings.TrimSpace(source[closeBracket+2 : closeParen])
	if len(destination) >= 2 && (destination[0] == '"' || destination[0] == '\'') &&
		destination[len(destination)-1] == destination[0] {
		destination = destination[1 : len(destination)-1]
	}
	fields := strings.Fields(destination)
	if len(fields) == 0 {
		return "", "", 0, false
	}
	destination = fields[0]
	return text, destination, closeParen + 1, true
}

// findMatching finds the matching delimiter for an opening delimiter.
func findMatching(source string, start int, open, close byte) int {
	depth := 0
	for i := start; i < len(source); i++ {
		if source[i] == '\\' {
			i++
			continue
		}
		switch source[i] {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// isEscapable reports whether a byte is escapable with a backslash.
func isEscapable(r byte) bool {
	return strings.ContainsRune(`!"#$%&'()*+,-./:;<=>?@[\]^_{|}~\`+"`", rune(r))
}

// escapeHTML escapes text for insertion into HTML text nodes.
func escapeHTML(source string) string {
	source = strings.ReplaceAll(source, "&", "&amp;")
	source = strings.ReplaceAll(source, "<", "&lt;")
	source = strings.ReplaceAll(source, ">", "&gt;")
	return source
}

// escapeHTMLAttribute escapes text for insertion into quoted HTML attributes.
func escapeHTMLAttribute(source string) string {
	source = escapeHTML(source)
	source = strings.ReplaceAll(source, `"`, "&quot;")
	source = strings.ReplaceAll(source, "'", "&#39;")
	return source
}

// isRawHTMLAt reports whether the source at start begins an HTML construct.
func isRawHTMLAt(source string, start int) bool {
	if start+1 >= len(source) {
		return false
	}
	next := source[start+1]
	return next == '/' || (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') ||
		next == '!' || next == '?'
}

// findRawHTMLEnd finds the end of a raw HTML tag or comment.
func findRawHTMLEnd(source string, start int) int {
	if strings.HasPrefix(source[start:], "<!--") {
		if end := strings.Index(source[start+4:], "-->"); end >= 0 {
			return start + 4 + end + 3
		}
	}
	if end := strings.IndexByte(source[start:], '>'); end >= 0 {
		return start + end + 1
	}
	return 0
}
