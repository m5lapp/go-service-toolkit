package jsonz

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// This package is called jsonz to avoid naming clashes with the standard
// library's encoding/json package. The suffix z was inspired by Google's adding
// of a z to common HTTP endpoints such as /healthz.

// Evelope represents the payload of a JSON response mapping string keys to
// any types of value.
type Envelope map[string]any

// WriteJSON marshals the contents of data into a JSON payload and writes it to
// the given http.ResponseWriter with the given HTTP status code and headers.
func WriteJSON(w http.ResponseWriter, status int, headers http.Header, data any) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// ReadJSON reads the JSON payload from a http.Request and decodes it into the
// given dst parameter, checking for any errors along the way.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	const maxBytes int64 = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	err := DecodeJSON(r.Body, dst, false)
	if err != nil {
		return err
	}

	return nil
}

// DecodeJSON unmarshals the contents of j into dst, checking for and returning
// any errors along the way. The unknownFields parameter determines whether
// having any fields in j that do not exist in dst causes an error to be
// returned.
func DecodeJSON(j io.Reader, dst any, unknownFields bool) error {
	dec := json.NewDecoder(j)

	if !unknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("JSON body contains badly formatted json at character %d", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("JSON body contains badly formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("JSON body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("JSON body contains incorrect JSON type at character %d", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("JSON body must not be empty")

		// Check there are no additional, unknown fields in the JSON body.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("JSON body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("JSON body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Call Decode again to check if there is any additional data in the string.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("JSON body must only contain a single JSON value")
	}

	return nil
}
