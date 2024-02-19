package main

/* Server configuration, routing and the like */

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/3WDeveloper-GM/library_app/backend/cli/handlers"
	"github.com/3WDeveloper-GM/library_app/backend/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func getApiRoutes(app *config.App) *chi.Mux {
	r := chi.NewMux()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.VisitedRouteLogger)
	r.NotFound(app.NotFoundResponse)
	r.MethodNotAllowed(app.MethodNotAllowedResponse)

	r.Post("/v1/insert/book", handlers.InsertEntryHandlerPost(app))
	r.Get("/v1/fetch/book/{id}", handlers.FetchEntryHandlerGet(app))
	r.Get("/v1/fetch/author/{id}", handlers.FetchAuthorEntryHandlerGet(app))
	r.Patch("/v1/update/book/{id}", handlers.UpdateEntriesHandlerPatch(app))
	r.Delete("/v1/delete/book/{id}", handlers.DeleteEntryHandlerDelete(app))

	return r
}

func StartServer(app *config.App) error {

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.ConfigFlags.Port),
		Handler:      getApiRoutes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	/* Graceful shutdown section */

	shutdownErr := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Log.Info().
			Str("signal", s.String()).
			Msg("shutting server down with signal")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr <- server.Shutdown(ctx)
	}()

	app.Log.Info().
		Interface("configuration", app.ConfigFlags).
		Msg("started server with the configuration")

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.Log.Info().Msg("server stopped.")

	return nil
}
