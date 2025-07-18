package har

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "test.har")
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return file
}

func TestLoadAndParse(t *testing.T) {
	validHAR := `{"log":{"entries":[{"request":{"method":"GET","url":"/foo"},"response":{"status":200,"headers":[],"content":{"text":"ok","mimeType":"text/plain"}}}]}}`
	badJSON := `{this is not json}`
	empty := ``
	missingFields := `{"log":{}}`
	extraFields := `{"log":{"entries":[{"request":{"method":"GET","url":"/foo"},"response":{"status":200,"headers":[],"content":{"text":"ok","mimeType":"text/plain"}},"extra":123}]}}`

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{"Valid HAR", validHAR, false},
		{"Bad JSON", badJSON, true},
		{"Empty File", empty, true},
		{"Missing Fields", missingFields, false}, // Should parse, but entries will be empty
		{"Extra Fields", extraFields, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file := writeTempFile(t, tc.content)
			_, err := LoadAndParse(file)
			if (err != nil) != tc.expectError {
				t.Errorf("LoadAndParse(%s) error = %v, expectError = %v", tc.name, err, tc.expectError)
			}
		})
	}
}
