package webapp

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
)

// clientErrorResponse is a struct that can be loaded into the data parameter of
// a JSend Fail response to provide the uclient with useful information about
// what went wrong. It should not be used for server-side errors.
type clientErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Action  string `json:"action,omitempty"`
}

// NewClientErrorResponse returns a new clientErrorResponse struct with the
// given values.
func NewClientErrorResponse(err, details, action string) clientErrorResponse {
	if err == "" {
		err = "No error details provided"
	}

	return clientErrorResponse{
		Error:   err,
		Details: details,
		Action:  action,
	}
}

// logError is a helper function for logging errors along with details of the
// request that caused it.
func (app *WebApp) logError(r *http.Request, err error) {
	trace := debug.Stack()
	app.Logger.Error(err.Error(),
		"request_method", r.Method,
		"request_url", r.URL.String(),
		"stack_trace", string(trace),
	)
}

// Server-side error response functions.

// errorResponse requests an arbitrary JSend-formatted HTTP response for a
// server-side error be sent to the client with the given HTTP status, message,
// local error code and data.
func (app *WebApp) errorResponse(w http.ResponseWriter, r *http.Request,
	status int, message string, code *int, data any) {
	err := jsonz.WriteJSendError(w, status, nil, message, code, &data)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ServerErrorResponse sends a generic error message to the client with an HTTP
// 500 (Internal Server Error) error code so as not to disclose to much
// information to the client. Details of the error will be logged locally.
func (app *WebApp) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "The server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, msg, nil, nil)
}

// Client-side error response functions.
// The convention used here is provide the client with a map which always
// contains an "error" parameter and optionally, "details" and "action"
// parameters with further information on how they can recover.

// FailResponse requests an arbitrary JSend-formatted HTTP response for a
// client-side error be sent to the client with the given HTTP status and data.
func (app *WebApp) FailResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	err := jsonz.WriteJSendFail(w, status, nil, data)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// NotFoundResponse returns an HTTP 404 (Not Found) response with an appropriate
// error message.
func (app *WebApp) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error": "The requested resource could not be found",
	}
	app.FailResponse(w, r, http.StatusNotFound, data)
}

// MethodNotAllowedError returns an HTTP 405 (Method Not Allowed) response with
// an appropriate error message and help text.
func (app *WebApp) MethodNotAllowedError(w http.ResponseWriter, r *http.Request) {
	e := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	a := "Check the Allow header of an OPTIONS request for a list of accepted methods"
	data := map[string]string{
		"error":  e,
		"action": a,
	}
	app.FailResponse(w, r, http.StatusMethodNotAllowed, data)
}

// BadRequestResponse returns an HTTP 400 (Bad Request) response with an
// appropriate error message.
func (app *WebApp) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	data := map[string]string{
		"error": err.Error(),
	}
	app.FailResponse(w, r, http.StatusBadRequest, data)
}

// FailedValidationResponse returns an HTTP 422 (Unprocessable Entity) response
// with an appropriate error message.
func (app *WebApp) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.FailResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// EditConflictResponse means that a resource could not be updated due to a
// recent update preventing it. An HTTP 409 (Conflict) response is returned with
// an appropriate error message.
func (app *WebApp) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "Unable to update the record due to an edit conflict",
		"action": "Please try again",
	}
	app.FailResponse(w, r, http.StatusConflict, data)
}

// RateLimitExceeded returns an HTTP 429 (Too Many Requests) response with an
// appropriate message and further help details.
func (app *WebApp) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":   "Rate limit exceeded",
		"details": "Large numbers of requests from the same IP address are limited over time",
		"action":  "Wait a while and then try again, or send fewer requests",
	}
	app.FailResponse(w, r, http.StatusTooManyRequests, data)
}

// InvalidCredentialsResponse returns an HTTP 401 (Unauthorized) response and an
// appropriate error message.
func (app *WebApp) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "Invalid authentication credentials",
		"action": "Check the credentials provided and that an account definitely exists",
	}
	app.FailResponse(w, r, http.StatusUnauthorized, data)
}

// InvalidAuthenticationTokenResponse returns an HTTP 401 (Unauthorized) reponse
// along with an appropriate error message and help text.
func (app *WebApp) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	data := map[string]string{
		"error":  "Invalid or missing authentication token",
		"action": "Check the token provided or request a new one and try again",
	}
	app.FailResponse(w, r, http.StatusUnauthorized, data)
}

// AuthenticationRequiredResponse returns an HTTP 401 (Unathorized) reponse along with an
// appropriate error message and help text.
func (app *WebApp) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "You must be authenticated to access this resource",
		"action": "Authenticate yourself, then try again with the provided credentials",
	}
	app.FailResponse(w, r, http.StatusUnauthorized, data)
}

// InactiveAccountResponse returns an HTTP 403 (Forbidden) reponse along with an
// appropriate error message and help text.
func (app *WebApp) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "Your user account must be activated to access this resource",
		"action": "Activate your account and then try again",
	}
	app.FailResponse(w, r, http.StatusForbidden, data)
}

// NotPermittedResponse returns an HTTP 403 (Forbidden) reponse along with an
// appropriate error message and help text.
func (app *WebApp) NotPermittedResponse(w http.ResponseWriter, r *http.Request) {
	e := "Your user account does not have the necessary permissions to access this resource"
	data := map[string]string{
		"error": e,
	}
	app.FailResponse(w, r, http.StatusForbidden, data)
}
