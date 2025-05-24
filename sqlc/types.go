package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// NewNullString creates a sql.NullString with the given value and sets the Valid flag to true if the string is not empty.
func NewNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

func NullStringToString(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}

// NewNullBool creates a sql.NullBool with the given value and sets the Valid flag to true.
func NewNullBool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: true,
	}
}

// NewNullTime creates a sql.NullTime with the given value and sets the Valid flag to true if the time is not the zero value.
func NewNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: t != time.Time{},
	}
}

// NewNullInt32 creates a sql.NullInt32 with the given value and sets the Valid flag to true.
func NewNullInt32(i int32) sql.NullInt32 {
	return sql.NullInt32{
		Int32: i,
		Valid: true,
	}
}

// NewNullInt16 creates a sql.NullInt16 with the given value and sets the Valid flag to true.
func NewNullInt16(i int16) sql.NullInt16 {
	return sql.NullInt16{
		Int16: i,
		Valid: true,
	}
}

// NewNullUUID creates a uuid.NullUUID with the given value and sets the Valid flag to true.
func NewNullUUID(u uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: u, Valid: true}
}

// NullTimeToDateOnlyString converts a sql.NullTime to a string in the format "2006-01-02" or an empty string if the time is not valid.
func NullTimeToDateOnlyString(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02")
}
func NullUUIDToString(u uuid.NullUUID) string {
	if !u.Valid {
		return ""
	}
	return u.UUID.String()
}
