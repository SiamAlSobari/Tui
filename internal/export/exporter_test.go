package export

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestToCSV(t *testing.T) {
	headers := []string{"id", "name", "note"}
	rows := [][]string{
		{"1", "Alice", "Hello, world"},
		{"2", "Bob", "He said, \"yes\""},
	}

	csv := ToCSV(headers, rows)
	lines := strings.Split(strings.TrimSpace(csv), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "id,name,note" {
		t.Errorf("expected header 'id,name,note', got %q", lines[0])
	}

	// Commas should cause fields to be double-quoted
	if lines[1] != `1,Alice,"Hello, world"` {
		t.Errorf("expected first row '1,Alice,\"Hello, world\"', got %q", lines[1])
	}

	// Quotes should be escaped by doubling them
	if lines[2] != `2,Bob,"He said, ""yes"""` {
		t.Errorf("expected second row '2,Bob,\"He said, \"\"yes\"\"\"', got %q", lines[2])
	}
}

func TestToJSON(t *testing.T) {
	headers := []string{"id", "name"}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}

	jsonStr := ToJSON(headers, rows)
	var data []map[string]string
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if len(data) != 2 {
		t.Fatalf("expected slice of size 2, got %d", len(data))
	}

	if data[0]["id"] != "1" || data[0]["name"] != "Alice" {
		t.Errorf("unexpected data in first element: %v", data[0])
	}
}

func TestToMarkdown(t *testing.T) {
	headers := []string{"id", "name"}
	rows := [][]string{
		{"1", "Alice"},
	}

	md := ToMarkdown(headers, rows)
	lines := strings.Split(strings.TrimSpace(md), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for markdown table, got %d", len(lines))
	}

	// Header line
	if !strings.Contains(lines[0], "id") || !strings.Contains(lines[0], "name") {
		t.Errorf("unexpected header line: %q", lines[0])
	}

	// Separator line
	if !strings.Contains(lines[1], "---") {
		t.Errorf("unexpected separator line: %q", lines[1])
	}

	// Data line
	if !strings.Contains(lines[2], "Alice") {
		t.Errorf("unexpected data line: %q", lines[2])
	}
}
