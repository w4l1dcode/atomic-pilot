package atomic

import (
	"fmt"
	"log"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func InitializeDB(dbPath string) (*sqlite.Conn, error) {
	conn, err := sqlite.OpenConn(dbPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := sqlitex.ExecScript(conn, `
		CREATE TABLE IF NOT EXISTS used_techniques (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			technique TEXT NOT NULL UNIQUE
		);
	`); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close connection after error: %w", closeErr)
		}
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return conn, nil
}

func CloseDB(conn *sqlite.Conn) {
	if err := conn.Close(); err != nil {
		log.Fatalf("Failed to close database connection: %v\n", err)
	}
}
