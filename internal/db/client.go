package db

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DBClient struct {
	DB   *sql.DB
	Path string
}

func OpenConnection(path string, readOnly bool) (*DBClient, error) {
	// 1. Validate file existence
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", path)
	}

	// 2. Validate SQLite 3 magic header
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for header validation: %w", err)
	}
	header := make([]byte, 16)
	n, err := io.ReadFull(f, header)
	f.Close()
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}
	if n < 16 || string(header[:16]) != "SQLite format 3\x00" {
		return nil, fmt.Errorf("invalid SQLite 3 database header")
	}

	var db *sql.DB
	var openErr error

	// If readOnly is requested, try to open in read-only mode first.
	if readOnly {
		dsn := fmt.Sprintf("file:%s?mode=ro", filepath.ToSlash(path))
		db, openErr = sql.Open("sqlite", dsn)
	} else {
		// Try read-write mode first.
		db, openErr = sql.Open("sqlite", path)
		if openErr == nil {
			// Ping to ensure connection is actually established and not locked
			openErr = db.Ping()
			if openErr != nil {
				// If read-write fails, fallback to read-only mode
				db.Close()
				dsn := fmt.Sprintf("file:%s?mode=ro", filepath.ToSlash(path))
				db, openErr = sql.Open("sqlite", dsn)
				if openErr == nil {
					openErr = db.Ping()
				}
			}
		}
	}

	if openErr != nil {
		if db != nil {
			db.Close()
		}
		return nil, fmt.Errorf("failed to open sqlite database: %w", openErr)
	}

	return &DBClient{
		DB:   db,
		Path: path,
	}, nil
}

func (c *DBClient) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
