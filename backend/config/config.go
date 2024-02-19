package config

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	books "github.com/3WDeveloper-GM/library_app/backend/internal/Books"
	"github.com/3WDeveloper-GM/library_app/backend/logger"
	"github.com/go-chi/chi/v5"

	_ "github.com/lib/pq"
)

type App struct {
	ConfigFlags struct {
		Port        int    `json:"port"`
		Environment string `json:"env"`
	}
	Database struct {
		DSN string
		DB  *sql.DB
	}
	Models struct {
		Create books.CreateEntryModel
		Read   books.ReadEntryModel
		Update books.UpdateEntryModel
		Delete books.DeleteEntryModel
	}
	logger.Logger
}

func NewAppObject() *App {
	return &App{}
}

func (app *App) SetConfigFlags() {

	flag.IntVar(&app.ConfigFlags.Port, "port", 8080, "Backend server port")
	flag.StringVar(&app.ConfigFlags.Environment, "env", "development", "environment (development|production|staging)")
	flag.StringVar(&app.Database.DSN, "dsn-db", os.Getenv("COCKROACHDB_DSN"), "CockroachDB database dsn")
	flag.Parse()
}

func (app *App) SetLogger() {
	app.Log = logger.NewLogger(app.ConfigFlags.Environment)
}

func (app *App) SetModels() {
	app.Models.Create.DB = app.Database.DB
	app.Models.Read.DB = app.Database.DB
	app.Models.Update.DB = app.Database.DB
	app.Models.Delete.DB = app.Database.DB
}

func (app *App) SetDB() error {

	db, err := app.OpenDB(app.Database.DSN)

	if err != nil {
		app.Log.Panic().Err(err).Send()
	}

	app.Database.DB = db

	return nil
}

func (app *App) ReadIDParams(r *http.Request) (int64, error) {
	id := chi.URLParam(r, "id")

	n, err := strconv.ParseInt(id, 10, 64)

	if err != nil || n < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return n, nil
}

func (app *App) OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", app.Database.DSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, err
}
