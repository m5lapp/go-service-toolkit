package webapp

import (
	"context"
	"errors"
	"expvar"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/m5lapp/go-service-toolkit/config"
	"golang.org/x/exp/slog"
)

// WebApp represents a generic, base web application or API that provides core
// functionality for receiving and responding to requests, logging, health
// checking, panic and error handling and various middlewares and functions.
type WebApp struct {
	ServerConfig config.Server
	Logger       *slog.Logger
	Router       *httprouter.Router
	Started      time.Time
	Wg           *sync.WaitGroup
}

// New returns a new WebApp with the given ServerConfig and Logger set.
func New(cfg config.Server, logger *slog.Logger) WebApp {
	wa := WebApp{
		ServerConfig: cfg,
		Logger:       logger,
		Router:       &httprouter.Router{},
		Started:      time.Now(),
		Wg:           &sync.WaitGroup{},
	}

	// Now that the WebApp is created, we can add the basic, common routes.
	wa.baseRoutes()

	return wa
}

// baseRoutes adds the standard routes that all WebApps should have.
func (app *WebApp) baseRoutes() {
	app.Router.MethodNotAllowed = http.HandlerFunc(app.MethodNotAllowedError)
	app.Router.NotFound = http.HandlerFunc(app.NotFoundResponse)

	app.Router.Handler(http.MethodGet, "/debug", expvar.Handler())
	app.Router.HandlerFunc(http.MethodGet, "/health", app.HealthCheckHandler)
}

// Serve configures an http.Server and starts it running whilst also spawning a
// goroutine to catch certain interrupt signals and handle them more gracefully.
func (app *WebApp) Serve(routes http.Handler) error {
	srv := &http.Server{
		Addr: app.ServerConfig.Addr,
		// TODO: Get this working with slog.
		//ErrorLog:     log.New(app.Logger, "", 0),
		Handler:      routes,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	// Start a background goroutine to catch shutdown signals.
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a
		// signal is received.
		s := <-quit

		app.Logger.Info("Shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Logger.Info("Completing background tasks", "addr", srv.Addr)

		app.Wg.Wait()
		shutdownError <- nil
	}()

	app.Logger.Info("Starting server", "env", app.ServerConfig.Env, "addr", srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Wait for the return value from the Shutdown() method.
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.Info("Stopped server", "addr", srv.Addr)
	return nil
}
