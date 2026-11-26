package render

import (
	"fmt"
	"strings"
)

type shortcodeCall struct {
	name string
	args map[string]string
}

func parseStandaloneShortcode(line string) (shortcodeCall, bool, error) {
	source := strings.TrimSpace(line)
	if !strings.HasPrefix(source, "{{<") {
		return shortcodeCall{}, false, nil
	}

	call, end, err := parseShortcodeAt(source, 0)
	if err != nil {
		return shortcodeCall{}, false, err
	}
	if strings.TrimSpace(source[end:]) != "" {
		return shortcodeCall{}, false, nil
	}
	return call, true, nil
}

func parseShortcodeAt(source string, start int) (shortcodeCall, int, error) {
	contentStart := start + len("{{<")
	relativeEnd := findShortcodeEnd(source[contentStart:])
	if relativeEnd < 0 {
		return shortcodeCall{}, 0, fmt.Errorf("shortcode is missing closing >}}")
	}
	contentEnd := contentStart + relativeEnd
	call, err := parseShortcodeContent(source[contentStart:contentEnd])
	if err != nil {
		return shortcodeCall{}, 0, err
	}
	return call, contentEnd + len(">}}"), nil
}

func findShortcodeEnd(source string) int {
	var quote byte
	escaped := false
	for i := 0; i+2 < len(source); i++ {
		char := source[i]
		if escaped {
			escaped = false
			continue
		}
		if quote != 0 {
			if char == '\\' {
				escaped = true
			} else if char == quote {
				quote = 0
			}
			continue
		}
		if char == '"' || char == '\'' {
			quote = char
			continue
		}
		if strings.HasPrefix(source[i:], ">}}") {
			return i
		}
	}
	return -1
}

func parseShortcodeContent(source string) (shortcodeCall, error) {
	source = strings.TrimSpace(source)
	nameEnd := 0
	for nameEnd < len(source) && !isShortcodeSpace(source[nameEnd]) {
		nameEnd++
	}
	if nameEnd == 0 {
		return shortcodeCall{}, fmt.Errorf("shortcode name is empty")
	}

	name := source[:nameEnd]
	if !validShortcodeName(name) {
		return shortcodeCall{}, fmt.Errorf("invalid shortcode name %q", name)
	}

	call := shortcodeCall{name: name, args: make(map[string]string)}
	for i := nameEnd; i < len(source); {
		i = skipShortcodeSpaces(source, i)
		if i >= len(source) {
			break
		}

		keyStart := i
		for i < len(source) && isShortcodeNameChar(source[i]) {
			i++
		}
		if i == keyStart {
			return shortcodeCall{}, fmt.Errorf("expected argument in key=value form")
		}
		key := source[keyStart:i]
		i = skipShortcodeSpaces(source, i)
		if i >= len(source) || source[i] != '=' {
			return shortcodeCall{}, fmt.Errorf("argument %q must use key=value form", key)
		}
		i++
		i = skipShortcodeSpaces(source, i)
		if i >= len(source) {
			return shortcodeCall{}, fmt.Errorf("argument %q has no value", key)
		}

		value, next, err := parseShortcodeValue(source, i)
		if err != nil {
			return shortcodeCall{}, fmt.Errorf("invalid value for argument %q: %w", key, err)
		}
		if _, exists := call.args[key]; exists {
			return shortcodeCall{}, fmt.Errorf("argument %q is specified more than once", key)
		}
		call.args[key] = value
		i = next
	}
	return call, nil
}

func parseShortcodeValue(source string, start int) (string, int, error) {
	if source[start] != '"' && source[start] != '\'' {
		end := start
		for end < len(source) && !isShortcodeSpace(source[end]) {
			end++
		}
		return source[start:end], end, nil
	}

	quote := source[start]
	var value strings.Builder
	for i := start + 1; i < len(source); i++ {
		if source[i] == '\\' && i+1 < len(source) {
			if source[i+1] == quote || source[i+1] == '\\' {
				value.WriteByte(source[i+1])
			} else {
				value.WriteByte(source[i])
				value.WriteByte(source[i+1])
			}
			i++
			continue
		}
		if source[i] == quote {
			next := i + 1
			if next < len(source) && !isShortcodeSpace(source[next]) {
				return "", 0, fmt.Errorf("unexpected text after quoted value")
			}
			return value.String(), next, nil
		}
		value.WriteByte(source[i])
	}
	return "", 0, fmt.Errorf("quoted value is not closed")
}

func validShortcodeName(name string) bool {
	if name == "" || !isShortcodeNameStart(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		if !isShortcodeNameChar(name[i]) {
			return false
		}
	}
	return true
}

func isShortcodeNameStart(char byte) bool {
	return char == '_' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

func isShortcodeNameChar(char byte) bool {
	return isShortcodeNameStart(char) || char >= '0' && char <= '9' || char == '-'
}

func skipShortcodeSpaces(source string, start int) int {
	for start < len(source) && isShortcodeSpace(source[start]) {
		start++
	}
	return start
}

func isShortcodeSpace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}

func (state *markdownRenderState) renderShortcode(call shortcodeCall) string {
	if state.executeShortcode == nil {
		state.err = fmt.Errorf("cannot render shortcode %q without a shortcode renderer", call.name)
		return ""
	}
	output, err := state.executeShortcode(call.name, call.args)
	if err != nil {
		state.err = fmt.Errorf("failed to render shortcode %q: %w", call.name, err)
		return ""
	}
	return string(output)
}
