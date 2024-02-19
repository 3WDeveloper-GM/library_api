package config

import (
	"fmt"
	"net/http"
)

func (app *App) ErrResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := Envelope{
		"error": message,
	}

	err := app.WriteJson(w, r, status, env, nil)
	if err != nil {
		app.Log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *App) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.Log.Error().Err(err).Send()

	message := "the server encountered a problem and could not process your request"
	app.ErrResponse(w, r, http.StatusInternalServerError, message)
}

func (app *App) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.ErrResponse(w, r, http.StatusNotFound, message)
}

func (app *App) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.ErrResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *App) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.ErrResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *App) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.ErrResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *App) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.ErrResponse(w, r, http.StatusConflict, message)
}
