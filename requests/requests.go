package requests

import (
	"bytes"
	"encoding/json"
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
func RequestJSend(method, url string, tOut time.Duration,
	body any, dataStruct any) (int, *jsonz.JSendResponse, error) {
	var reqBody io.Reader

	// If a request body has been provided, attempt to marshal it into JSON.
	if body != nil {
		js, err := json.Marshal(body)
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

	dst := &jsonz.JSendResponse{Data: dataStruct}
	err = jsonz.DecodeJSON(res.Body, dst)
	if err != nil {
		return 0, nil, err
	}

	return res.StatusCode, dst, nil
}
