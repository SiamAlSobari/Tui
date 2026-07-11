package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
)

func ToCSV(headers []string, rows [][]string) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	// Write header
	if err := w.Write(headers); err != nil {
		return ""
	}
	// Write rows
	if err := w.WriteAll(rows); err != nil {
		return ""
	}
	w.Flush()
	return buf.String()
}

func ToJSON(headers []string, rows [][]string) string {
	var data []map[string]string
	for _, row := range rows {
		m := make(map[string]string)
		for i, h := range headers {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			m[h] = val
		}
		data = append(data, m)
	}

	bytesVal, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(bytesVal)
}

func ToMarkdown(headers []string, rows [][]string) string {
	var s strings.Builder

	// Header row
	s.WriteString("| ")
	s.WriteString(strings.Join(headers, " | "))
	s.WriteString(" |\n")

	// Separator row
	s.WriteString("|")
	for range headers {
		s.WriteString(" --- |")
	}
	s.WriteString("\n")

	// Data rows
	for _, row := range rows {
		s.WriteString("| ")
		rowVals := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				rowVals[i] = row[i]
			}
		}
		s.WriteString(strings.Join(rowVals, " | "))
		s.WriteString(" |\n")
	}

	return s.String()
}
