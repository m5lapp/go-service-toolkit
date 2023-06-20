package webapp

import (
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/m5lapp/go-service-toolkit/config"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// RecoverPanic recovers any panics that happen in the goroutine that handles
// the request. The defered function will close the connection, then print a log
// message including a full stack trace. It's important to note that if the
// goroutine creates any further goroutines, then these must handle any panics
// themselves.
func (app *WebApp) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				app.ServerErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RateLimit is a middleware function that limits the number of requests a
// client (based on their IP address) can make in a given period.
func (app *WebApp) RateLimit(cfg config.Limiter, next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a goroutine to clean up old entries from the client map every
	// minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Active {
			ip := realip.FromRequest(r)

			mu.Lock()

			_, found := clients[ip]
			if !found {
				r := rate.Limit(cfg.RPS)
				b := cfg.Burst
				clients[ip] = &client{limiter: rate.NewLimiter(r, b)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.RateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// EnableCORS is a middleware function that handles CORS (Cross-Origin Resource
// Sharing) requests to prmit a web browser to make requests to a different
// origin (domain, scheme or port) to the main we page.
func (app *WebApp) EnableCORS(cfg config.Cors, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for _, trustedOrigin := range cfg.TrustedOrigins {
				if origin == trustedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS, PATCH, PUT")

						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !mw.headerWritten {
		mw.statusCode = http.StatusOK
		mw.headerWritten = true
	}

	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

// Metrics is a middleware function that keeps track of a number of metrics
// relating to HTTP requests.
func (app *WebApp) Metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived  = expvar.NewInt("total_requests_received")
		totalResponsesSent     = expvar.NewInt("total_responses_sent")
		totalProcoessingTimeμs = expvar.NewInt("total_processing_times_μs")
		totalResponsesByStatus = expvar.NewMap("total_responses_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)

		mw := &metricsResponseWriter{wrapped: w}

		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)

		totalResponsesByStatus.Add(strconv.Itoa(mw.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcoessingTimeμs.Add(duration)
	})
}
