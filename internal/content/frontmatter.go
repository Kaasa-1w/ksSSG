package content

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type FrontMatter struct {
	Title string `toml:"title"`
	Date  string `toml:"date"`
	Draft bool   `toml:"draft"`
}

// ParseFrontMatter extracts TOML metadata and returns the remaining Markdown body.
func ParseFrontMatter(source []byte) (FrontMatter, []byte, error) {
	var fm FrontMatter
	var body []byte

	if !bytes.HasPrefix(source, []byte("+++\n")) && !bytes.HasPrefix(source, []byte("+++\r\n")) {
		return fm, body, errors.New("front matter is missing")
	}

	normalized := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))

	rest := normalized[len("+++\n"):]

	end := bytes.Index(rest, []byte("\n+++\n"))
	if end < 0 {
		if bytes.HasSuffix(rest, []byte("\n+++")) {
			end = len(rest) - len("\n+++")
			fm, err := parseTOMLFrontMatter(rest[:end])
			if err != nil {
				return fm, body, err
			}
			return fm, body, nil
		}

		return fm, body, errors.New("front matter is not closed")
	}

	fm, err := parseTOMLFrontMatter(rest[:end])
	if err != nil {
		return fm, body, err
	}

	body = rest[end+len("\n+++\n"):]

	return fm, body, nil
}

// parseTOMLFrontMatter parses metadata into FrontMatter using its TOML field tags.
func parseTOMLFrontMatter(source []byte) (FrontMatter, error) {
	var fm FrontMatter
	if err := parseTOMLFields(source, &fm); err != nil {
		return fm, err
	}
	return fm, nil
}

// parseTOMLFields parses TOML key-value pairs into a struct using its field tags.
func parseTOMLFields(source []byte, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() || targetValue.Elem().Kind() != reflect.Struct {
		return errors.New("front matter target must be a non-nil pointer to a struct")
	}
	targetValue = targetValue.Elem()
	typeOf := targetValue.Type()

	fields := make(map[string]int, typeOf.NumField())
	for index := 0; index < typeOf.NumField(); index++ {
		field := typeOf.Field(index)
		if field.PkgPath != "" {
			continue
		}
		name := strings.Split(field.Tag.Get("toml"), ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		fields[name] = index
	}

	seen := make(map[string]bool)

	for lineNumber, line := range bytes.Split(source, []byte("\n")) {
		text := strings.TrimSpace(string(line))
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		key, rawValue, ok := strings.Cut(text, "=")
		if !ok {
			return fmt.Errorf("failed to parse front matter TOML: line %d: expected key = value", lineNumber+1)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("failed to parse front matter TOML: line %d: key is empty", lineNumber+1)
		}
		if seen[key] {
			return fmt.Errorf("failed to parse front matter TOML: line %d: duplicate key %q", lineNumber+1, key)
		}
		seen[key] = true

		rawValue, err := stripTOMLComment(strings.TrimSpace(rawValue))
		if err != nil {
			return fmt.Errorf("failed to parse front matter TOML: line %d: %w", lineNumber+1, err)
		}

		fieldIndex, ok := fields[key]
		if !ok {
			return fmt.Errorf("failed to parse front matter TOML: line %d: unsupported key %q", lineNumber+1, key)
		}
		if err := setTOMLField(targetValue.Field(fieldIndex), rawValue, key); err != nil {
			return fmt.Errorf("failed to parse front matter TOML: line %d: %s: %w", lineNumber+1, key, err)
		}
	}
	return nil
}

// setTOMLField parses a scalar value according to the destination field type.
func setTOMLField(field reflect.Value, rawValue, key string) error {
	switch field.Kind() {
	case reflect.String:
		value, err := parseTOMLString(rawValue)
		if err != nil {
			return err
		}
		field.SetString(value)
	case reflect.Bool:
		if rawValue != "true" && rawValue != "false" {
			return fmt.Errorf("%s must be true or false", key)
		}
		field.SetBool(rawValue == "true")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(rawValue, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("%s must be an integer: %w", key, err)
		}
		field.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(rawValue, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("%s must be an unsigned integer: %w", key, err)
		}
		field.SetUint(value)
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(rawValue, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("%s must be a number: %w", key, err)
		}
		field.SetFloat(value)
	default:
		return fmt.Errorf("%s has unsupported field type %s", key, field.Type())
	}
	return nil
}

// stripTOMLComment removes a comment that appears outside a quoted string.
func stripTOMLComment(value string) (string, error) {
	var quote rune
	escaped := false
	for index, char := range value {
		if escaped {
			escaped = false
			continue
		}
		if quote == '"' && char == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		switch char {
		case '"', '\'':
			quote = char
		case '#':
			return strings.TrimSpace(value[:index]), nil
		}
	}
	if quote != 0 || escaped {
		return "", errors.New("unterminated quoted string")
	}
	return strings.TrimSpace(value), nil
}

// parseTOMLString decodes a single-line TOML basic or literal string.
func parseTOMLString(value string) (string, error) {
	if len(value) < 2 {
		return "", errors.New("expected a quoted string")
	}
	quote := value[0]
	if quote != '"' && quote != '\'' {
		return "", errors.New("expected a quoted string")
	}
	if value[len(value)-1] != quote {
		return "", errors.New("unterminated quoted string")
	}

	if quote == '\'' {
		if strings.ContainsRune(value[1:len(value)-1], '\'') {
			return "", errors.New("literal string cannot contain a single quote")
		}
		return value[1 : len(value)-1], nil
	}

	for index := 1; index < len(value)-1; index++ {
		if value[index] == '\\' {
			index++
			if index >= len(value)-1 || !strings.ContainsRune("btnfr\"\\uU", rune(value[index])) {
				return "", errors.New("invalid escape sequence")
			}
		}
	}
	parsed, err := strconv.Unquote(value)
	if err != nil {
		return "", err
	}
	return parsed, nil
}
