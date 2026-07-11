package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenConnection_NonExistent(t *testing.T) {
	_, err := OpenConnection("non_existent_file_12345.db", false)
	if err == nil {
		t.Fatal("expected error when opening non-existent file, got nil")
	}
}

func TestOpenConnection_InvalidHeader(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.db")
	err := os.WriteFile(invalidFile, []byte("NOT A SQLITE HEADER"), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err = OpenConnection(invalidFile, false)
	if err == nil {
		t.Fatal("expected error when opening file with invalid sqlite3 header, got nil")
	}
}

func TestDatabase_Operations(t *testing.T) {
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.db")

	// 1. Create a valid SQLite database with a table and view
	// We must create it first because OpenConnection validates file existence and headers.
	initDB, err := sql.Open("sqlite", dbFile)
	if err != nil {
		t.Fatalf("failed to init temp db: %v", err)
	}
	_, err = initDB.Exec("PRAGMA journal_mode=WAL;") // Forces creation of the database file on disk
	if err != nil {
		initDB.Close()
		t.Fatalf("failed to exec schema pragma: %v", err)
	}
	initDB.Close()

	client, err := OpenConnection(dbFile, false)
	if err != nil {
		t.Fatalf("failed to open/create test db: %v", err)
	}
	defer client.Close()

	// Create test tables and view
	_, err = client.DB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			active INTEGER DEFAULT 1
		);
		CREATE VIEW active_users AS SELECT * FROM users WHERE active = 1;
		INSERT INTO users (email, active) VALUES ('alice@example.com', 1), ('bob@example.com', 0);
	`)
	if err != nil {
		t.Fatalf("failed to setup test database data: %v", err)
	}

	// 2. Test ListTables
	tables, err := ListTables(client)
	if err != nil {
		t.Fatalf("failed to list tables: %v", err)
	}

	// Expect two items: users (table, rowcount 2) and active_users (view, rowcount 1)
	if len(tables) != 2 {
		t.Errorf("expected 2 tables/views, got %d", len(tables))
	}

	var foundTable, foundView bool
	for _, info := range tables {
		if info.Name == "users" {
			foundTable = true
			if info.Type != "table" {
				t.Errorf("expected type 'table' for users, got %s", info.Type)
			}
			if info.RowCount != 2 {
				t.Errorf("expected row count 2 for users, got %d", info.RowCount)
			}
		} else if info.Name == "active_users" {
			foundView = true
			if info.Type != "view" {
				t.Errorf("expected type 'view' for active_users, got %s", info.Type)
			}
			if info.RowCount != 1 {
				t.Errorf("expected row count 1 for active_users, got %d", info.RowCount)
			}
		}
	}

	if !foundTable {
		t.Error("users table not found in ListTables")
	}
	if !foundView {
		t.Error("active_users view not found in ListTables")
	}

	// 3. Test GetTableSchema for table "users"
	cols, ddl, err := GetTableSchema(client, "users")
	if err != nil {
		t.Fatalf("failed to get table schema: %v", err)
	}

	if len(cols) != 3 {
		t.Fatalf("expected 3 columns for users, got %d", len(cols))
	}

	// Verify columns: id, email, active
	colMap := make(map[string]ColumnInfo)
	for _, col := range cols {
		colMap[col.Name] = col
	}

	idCol, ok := colMap["id"]
	if !ok {
		t.Error("column id not found")
	} else {
		if idCol.Type != "INTEGER" {
			t.Errorf("expected id type INTEGER, got %s", idCol.Type)
		}
		if !idCol.IsPK {
			t.Error("expected id to be primary key")
		}
	}

	emailCol, ok := colMap["email"]
	if !ok {
		t.Error("column email not found")
	} else {
		if emailCol.Type != "TEXT" {
			t.Errorf("expected email type TEXT, got %s", emailCol.Type)
		}
		if !emailCol.NotNull {
			t.Error("expected email to be NOT NULL")
		}
	}

	activeCol, ok := colMap["active"]
	if !ok {
		t.Error("column active not found")
	} else {
		if activeCol.DefaultVal != "1" {
			t.Errorf("expected active default '1', got %q", activeCol.DefaultVal)
		}
	}

	if ddl == "" {
		t.Error("expected non-empty DDL statement")
	}
}
