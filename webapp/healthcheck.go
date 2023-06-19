package webapp

import (
	"net/http"
	"time"

	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/vcs"
)

// HealthCheckHandler provides a basic health check response.
func (app *WebApp) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := jsonz.Envelope{
		"status": "available",
		"system_info": map[string]any{
			"environment": app.ServerConfig.Env,
			"started":     app.Started,
			"uptime":      time.Since(app.Started).String(),
			"version":     vcs.Version(),
		},
	}

	err := jsonz.WriteJSendSuccess(w, http.StatusOK, nil, data)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
