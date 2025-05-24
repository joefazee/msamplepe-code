package db

import (
	"database/sql"
)

type DBLogWriter struct {
	db *sql.DB
}

// NewDBLogWriter creates a new DBLogWriter.
func NewDBLogWriter(db *sql.DB) *DBLogWriter {
	return &DBLogWriter{db: db}
}

// Write implements the io.Writer interface.
func (w *DBLogWriter) Write(p []byte) (n int, err error) {
	jsonData := string(p)

	query := "INSERT INTO logs(log_data) VALUES($1::jsonb)"
	_, err = w.db.Exec(query, jsonData)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
