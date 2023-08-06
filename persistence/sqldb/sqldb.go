package sqldb

import (
	"context"
	"database/sql"
	"time"

	"github.com/m5lapp/go-service-toolkit/config"
)

// ErrUniqueConstraintViolation is a database-agnostic error for unique
// constraint violations. Columns is a slice of strings in case the constraints
// spans multiple columns.
type ErrUniqueConstraintViolation struct {
	Columns []string
	Message string
	Table   string
}

// Error implements the Error() interface.
func (e ErrUniqueConstraintViolation) Error() string {
	return e.Message
}

// Unwrap returns the base error that ErrUniqueConstraintViolationWrapped wraps.
// func (e *ErrUniqueConstraintViolationWrapped) Unwrap() error {
// 	return e.baseErr
// }

// NewUniqueConstraintErr returns a new ErrUniqueConstraintViolation with the
// fields appropriately set. At least one column must be provided and more can
// be provided in the case of composite unique constraints.
func NewUniqueConstraintErr(table, column string, columns ...string) *ErrUniqueConstraintViolation {
	cols := []string{column}
	cols = append(cols, columns...)

	msg := "a record already exists for this field"
	if len(cols) > 1 {
		msg = "a record already exists for these fields"
	}

	return &ErrUniqueConstraintViolation{
		Columns: cols,
		Message: msg,
		Table:   table,
	}
}

// OpenDB takes a config.SqlDB configuration, attempts to open a connection to
// it, sets some important properties and then tests the connection before
// returning a pointer to the connection and an error.
func OpenDB(cfg config.SqlDB) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
