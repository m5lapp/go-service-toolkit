package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
)

// RequestJSend sends an HTTP request of the given method type to the given URL
// with the given JSON body and attempts to decode the response payload into a
// JSendResponse struct where the data field is the dataStruct struct. The
// JSendResponse struct is then returned along with the HTTP status code and an
// error.
func RequestJSend(method, url string, tOut time.Duration, requestBody any,
	targetStruct any) (int, *jsonz.JSendResponseRaw, error) {
	var reqBody io.Reader

	// If a request body has been provided, attempt to marshal it into JSON.
	if requestBody != nil {
		js, err := json.Marshal(requestBody)
		if err != nil {
			return 0, nil, err
		}

		reqBody = bytes.NewReader(js)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: tOut}
	res, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	// First pass to get the JSend response status.
	responseBody := &jsonz.JSendResponseRaw{}
	err = jsonz.DecodeJSON(res.Body, responseBody, true)
	if err != nil {
		return 0, nil, err
	}

	if responseBody.Status != jsonz.JSendStatusSuccess {
		switch responseBody.Status {
		case jsonz.JSendStatusError:
			return 0, nil, errors.New("todo error")
		case jsonz.JSendStatusFail:
			return 0, nil, errors.New("todo also error")
		default:
			return 0, nil, errors.New("invalid JSend status")
		}
	}

	// Second pass to decode the JSend response into the target struct.
	err = jsonz.DecodeJSON(bytes.NewReader(responseBody.Data), targetStruct, true)
	if err != nil {
		return 0, nil, err
	}

	return res.StatusCode, responseBody, nil
}
