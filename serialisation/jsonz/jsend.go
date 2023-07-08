package jsonz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	errorMsgStatus = "http status code (%d) for %s is not in range %d to %d"

	JSendStatusError   = "error"
	JSendStatusFail    = "fail"
	JSendStatusSuccess = "success"
)

// JSend Statuses should only be one of the JSendStatus* consts.
var ErrInvalidJSendStatus error = errors.New("invalid JSend status field")

func validJSendStatus(s string) bool {
	return s == JSendStatusError || s == JSendStatusFail || s == JSendStatusSuccess
}

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

// JSendResponseRaw also represents a JSend JSON response payload like
// JSendResponse, except the Data field is a json.RawMessage. This is useful for
// unmarshaling JSend responses into where the format of Data is unknown until
// the Status field has been decoded and checked.
type JSendResponseRaw struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Code    *int            `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
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

// RequestJSend sends an HTTP request of the given method type to the given URL
// with the given JSON requestBody and timeout and attempts to decode the
// response payload into a JSendResponseRaw struct.
//
// The HTTP response is returned along with the decoded JSendResponseRaw, and an
// error if there is one. The requesting function is then responsible for
// decoding the JSendResponseRaw.Data field if required.
func RequestJSend(method, url string, tOut time.Duration, requestBody any,
) (*http.Response, *JSendResponseRaw, error) {
	var reqBody io.Reader

	// If a request body has been provided, attempt to marshal it into JSON.
	if requestBody != nil {
		js, err := json.Marshal(requestBody)
		if err != nil {
			return nil, nil, err
		}

		reqBody = bytes.NewReader(js)
	}

	// Create and prepare the HTTP request.
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request.
	client := http.Client{Timeout: tOut}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// Decode the response body to get the JSend response status.
	jSendBody := &JSendResponseRaw{}
	err = DecodeJSON(httpResp.Body, jSendBody, true)
	if err != nil {
		return httpResp, nil, err
	}

	// Check the JSend response's Status field is valid.
	if !validJSendStatus(jSendBody.Status) {
		return httpResp, jSendBody, ErrInvalidJSendStatus
	}

	return httpResp, jSendBody, nil
}
