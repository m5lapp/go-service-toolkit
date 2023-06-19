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
// 500 error code so as not to disclose to much information to the client.
// Details of the error will be logged locally.
func (app *WebApp) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "The server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, msg, nil, nil)
}

// Client-side error response functions.
// The convention used here is provide the client with a map which always
// contains an "error" parameter and optionally, "details" and "action"
// parameters with further information on how they can recover.

// failResponse requests an arbitrary JSend-formatted HTTP response for a
// client-side error be sent to the client with the given HTTP status and data.
func (app *WebApp) failResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	err := jsonz.WriteJSendFail(w, status, nil, data)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *WebApp) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error": "The requested resource could not be found",
	}
	app.failResponse(w, r, http.StatusNotFound, data)
}

func (app *WebApp) MethodNotAllowedError(w http.ResponseWriter, r *http.Request) {
	e := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	a := "Check the Allow header of an OPTIONS request for a list of accepted methods"
	data := map[string]string{
		"error":  e,
		"action": a,
	}
	app.failResponse(w, r, http.StatusMethodNotAllowed, data)
}

func (app *WebApp) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	data := map[string]string{
		"error": err.Error(),
	}
	app.failResponse(w, r, http.StatusBadRequest, data)
}

func (app *WebApp) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.failResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *WebApp) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "Unable to update the record due to an edit conflict",
		"action": "Please try again",
	}
	app.failResponse(w, r, http.StatusConflict, data)
}

func (app *WebApp) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":   "Rate limit exceeded",
		"details": "Large numbers of requests from the same IP address are limited over time",
		"action":  "Wait a while and then try again, or send fewer requests",
	}
	app.failResponse(w, r, http.StatusTooManyRequests, data)
}

func (app *WebApp) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "Invalid authentication credentials",
		"action": "Check the credentials provided and that an account definitely exists",
	}
	app.failResponse(w, r, http.StatusUnauthorized, data)
}

func (app *WebApp) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	data := map[string]string{
		"error":  "Invalid or missing authentication token",
		"action": "Check the token provided or request a new one and try again",
	}
	app.failResponse(w, r, http.StatusUnauthorized, data)
}

func (app *WebApp) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "You must be authenticated to access this resource",
		"action": "Authenticate yourself, then try again with the provided credentials",
	}
	app.failResponse(w, r, http.StatusUnauthorized, data)
}

func (app *WebApp) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"error":  "Your user account must be activated to access this resource",
		"action": "Activate your account and then try again",
	}
	app.failResponse(w, r, http.StatusForbidden, data)
}

func (app *WebApp) NotPermittedResponse(w http.ResponseWriter, r *http.Request) {
	e := "Your user account does not have the necessary permissions to access this resource"
	data := map[string]string{
		"error": e,
	}
	app.failResponse(w, r, http.StatusForbidden, data)
}
