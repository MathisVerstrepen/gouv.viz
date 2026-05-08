package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func walkRawJSON(dir string, handle func(rawFile) error) error {
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("open raw directory %s: %w", dir, err)
	}

	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var root map[string]any
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}

		sourcePath := filepath.ToSlash(path)
		if rel, err := filepath.Rel(".", path); err == nil {
			sourcePath = filepath.ToSlash(rel)
		}

		hash := sha256.Sum256(data)
		return handle(rawFile{
			Path:       path,
			SourcePath: sourcePath,
			SourceHash: hex.EncodeToString(hash[:]),
			Root:       root,
		})
	})
}

func objectAt(root map[string]any, path ...string) map[string]any {
	return asObject(valueAt(root, path...))
}

func valueAt(root map[string]any, path ...string) any {
	var current any = root
	for _, key := range path {
		object := asObject(current)
		if object == nil {
			return nil
		}
		current = object[key]
	}
	return current
}

func stringAt(root map[string]any, path ...string) string {
	value, _ := stringValue(valueAt(root, path...))
	return value
}

func intAt(root map[string]any, path ...string) (int, bool) {
	return intValue(valueAt(root, path...))
}

func boolIntAt(root map[string]any, path ...string) (int, bool) {
	value := valueAt(root, path...)
	switch typed := value.(type) {
	case bool:
		if typed {
			return 1, true
		}
		return 0, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "oui", "yes":
			return 1, true
		case "false", "0", "non", "no":
			return 0, true
		}
	}
	return intValue(value)
}

func asObject(value any) map[string]any {
	if object, ok := value.(map[string]any); ok && !isNilObject(object) {
		return object
	}
	return nil
}

func items(value any) []any {
	if value == nil {
		return nil
	}
	if object, ok := value.(map[string]any); ok && isNilObject(object) {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		return typed
	case map[string]any:
		return []any{typed}
	default:
		if text, ok := stringValue(typed); ok && text != "" {
			return []any{text}
		}
	}
	return nil
}

func stringsFromValue(value any) []string {
	var values []string
	for _, item := range items(value) {
		if text, ok := stringValue(item); ok && text != "" {
			values = append(values, text)
		}
	}
	return values
}

func stringValue(value any) (string, bool) {
	if value == nil {
		return "", false
	}

	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		return text, text != ""
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(typed), true
	case map[string]any:
		if isNilObject(typed) {
			return "", false
		}
		return stringValue(typed["#text"])
	}

	return "", false
}

func intValue(value any) (int, bool) {
	if value == nil {
		return 0, false
	}

	switch typed := value.(type) {
	case int:
		return typed, true
	case float64:
		return int(typed), true
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return 0, false
		}
		parsed, err := strconv.Atoi(text)
		return parsed, err == nil
	case map[string]any:
		if isNilObject(typed) {
			return 0, false
		}
		return intValue(typed["#text"])
	}

	return 0, false
}

func isNilObject(value map[string]any) bool {
	if raw, ok := value["@xsi:nil"]; ok {
		text, _ := stringValue(raw)
		return text == "true" || text == "1"
	}
	return false
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullInt(value int, ok bool) any {
	if !ok {
		return nil
	}
	return value
}
