package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableInfo struct {
	Name     string // Table/View name
	Type     string // "table" or "view"
	RowCount int    // Total number of rows
}

func escapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\"\"")
}

func ListTables(client *DBClient) ([]TableInfo, error) {
	// Query sqlite_schema for user-defined tables and views
	query := `SELECT type, name FROM sqlite_schema WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY name ASC`
	rows, err := client.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite_schema: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var tType, name string
		if err := rows.Scan(&tType, &name); err != nil {
			return nil, fmt.Errorf("failed to scan table metadata: %w", err)
		}

		// Query dynamic row count for each table/view
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, escapeDoubleQuotes(name))
		var count int
		err := client.DB.QueryRow(countQuery).Scan(&count)
		if err != nil {
			// If counting fails (e.g. a view with invalid reference), default to 0
			count = 0
		}

		tables = append(tables, TableInfo{
			Name:     name,
			Type:     tType,
			RowCount: count,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table metadata: %w", err)
	}

	return tables, nil
}

type ColumnInfo struct {
	Name       string
	Type       string
	NotNull    bool
	DefaultVal string
	IsPK       bool
}

func GetTableSchema(client *DBClient, tableName string) ([]ColumnInfo, string, error) {
	// 1. Get column details via PRAGMA table_info
	// Pragma query cannot be parameterized directly with ?, so we escape the table name safely.
	pragmaQuery := fmt.Sprintf(`PRAGMA table_info("%s")`, escapeDoubleQuotes(tableName))
	rows, err := client.DB.Query(pragmaQuery)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query table_info for %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var cid int
		var name, colType string
		var notnull int
		var dfltVal sql.NullString
		var pk int

		err := rows.Scan(&cid, &name, &colType, &notnull, &dfltVal, &pk)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan column info: %w", err)
		}

		defaultValStr := ""
		if dfltVal.Valid {
			defaultValStr = dfltVal.String
		}

		columns = append(columns, ColumnInfo{
			Name:       name,
			Type:       colType,
			NotNull:    notnull == 1,
			DefaultVal: defaultValStr,
			IsPK:       pk > 0,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("error iterating column info: %w", err)
	}

	// 2. Retrieve DDL schema from sqlite_schema
	var sqlDDL sql.NullString
	err = client.DB.QueryRow(`SELECT sql FROM sqlite_schema WHERE type IN ('table', 'view') AND name = ?`, tableName).Scan(&sqlDDL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("table or view %q not found", tableName)
		}
		return nil, "", fmt.Errorf("failed to query DDL for %s: %w", tableName, err)
	}

	ddl := ""
	if sqlDDL.Valid {
		ddl = sqlDDL.String
	}

	return columns, ddl, nil
}
