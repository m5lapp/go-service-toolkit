package jsonz

import (
	"fmt"
	"net/http"
)

const (
	errorMsgStatus = "http status code (%d) for %s is not in range %d to %d"

	JSendStatusError   = "error"
	JSendStatusFail    = "fail"
	JSendStatusSuccess = "success"
)

// JSendResponse represents a JSend JSON response payload. The Status field is
// mandatory for all responses, whilst the others are used depending on whether
// the response is a success, a fail or an error. See:
// https://github.com/omniti-labs/jsend
type JSendResponse struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewJSendSuccess returns a JSendResponse that indicates that everything went
// well with the request and usually some data is also returned. See:
// https://github.com/omniti-labs/jsend#success
func NewJSendSuccess(data any) JSendResponse {
	return JSendResponse{
		Status: JSendStatusSuccess,
		Data:   data,
	}
}

// NewJSendFail returns a JSendResponse that indicates that the request was
// "rejected due to invalid data or call conditions". The data parameter should
// provide "details of why the request failed. If the reasons for failure
// correspond to POST values, the response object's keys SHOULD correspond to
// those POST values". See: https://github.com/omniti-labs/jsend#fail
func NewJSendFail(data any) JSendResponse {
	return JSendResponse{
		Status: JSendStatusFail,
		Data:   data,
	}
}

// NewJSendError returns a JSendResponse that indicates that the request failed
// due to an error on the server. The message should be meaningful to the
// end-user and explain what went wrong. Code is an optional numeric code for
// the error and data may contain additional information about the error such as
// a stack trace. See: https://github.com/omniti-labs/jsend#error
func NewJSendError(message string, code *int, data *any) JSendResponse {
	return JSendResponse{
		Status:  JSendStatusError,
		Message: message,
		Code:    code,
		Data:    data,
	}
}

// WriteJSendSuccess checks the provided HTTP status code is a valid success
// code, then calls WriteJSON with an appropriate JSend payload.
func WriteJSendSuccess(w http.ResponseWriter, status int, headers http.Header, data any) error {
	if status < 200 || status > 299 {
		return fmt.Errorf(errorMsgStatus, status, JSendStatusSuccess, 200, 299)
	}

	err := WriteJSON(w, status, headers, NewJSendSuccess(data))
	if err != nil {
		return err
	}

	return nil
}

// WriteJSendFail checks the provided HTTP status code is a valid client-side
// error code, then calls WriteJSON with an appropriate JSend payload.
func WriteJSendFail(w http.ResponseWriter, status int, headers http.Header, data any) error {
	if status < 400 || status > 499 {
		return fmt.Errorf(errorMsgStatus, status, JSendStatusFail, 400, 499)
	}

	err := WriteJSON(w, status, headers, NewJSendFail(data))
	if err != nil {
		return err
	}

	return nil
}

// WriteJSendError checks the provided HTTP status code is a valid server-side
// error code, then calls WriteJSON with an appropriate JSend payload.
func WriteJSendError(w http.ResponseWriter, status int, headers http.Header,
	message string, code *int, data *any) error {
	if status < 500 || status > 599 {
		return fmt.Errorf(errorMsgStatus, status, JSendStatusError, 500, 599)
	}

	err := WriteJSON(w, status, headers, NewJSendError(message, code, data))
	if err != nil {
		return err
	}

	return nil
}
