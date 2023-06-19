package jsonz

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

// DateOnly embedds a time.Time struct to represent a date without a time
// component.
type DateOnly struct {
	time.Time
}

var ErrInvalidDateOnlyFormat = errors.New("invalid date format for DateOnly")

// MarshalJSON implements the encoding/json.Marshaler interface to convert a
// DateOnly struct into a suitable JSON representation.
func (d DateOnly) MarshalJSON() ([]byte, error) {
	jsonValue := strconv.Quote(d.Format(time.DateOnly))
	return []byte(jsonValue), nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface to populate
// the DateOnly struct's value from a JSON representation.
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return ErrInvalidDateOnlyFormat
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}

	*d = DateOnly{Time: t}
	return nil
}

// Scan implements the database/sql.Scanner interface. It takes a values from
// the database (hopefully a string in the time.DateOnly format) and attempts to
// store it into the DateOnly struct.
func (d *DateOnly) Scan(value any) error {
	if value == nil {
		return nil
	}

	d.Time = value.(time.Time)

	return nil
}

// Value implements the database/sql/driver.Valuer interface. It returns a value
// of type time.Time from the DateOnly struct suitable for storing in a SQL
// database.
func (d DateOnly) Value() (driver.Value, error) {
	val, err := json.Marshal(d.Time)
	if err != nil {
		return nil, err
	}

	return val, nil
}
