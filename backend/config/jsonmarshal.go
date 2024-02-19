package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Envelope map[string]interface{}

func (app *App) WriteJson(w http.ResponseWriter, r *http.Request, status int, data Envelope, headers http.Header) error {

	var js []byte
	var err error

	if app.ConfigFlags.Environment != "development" {
		js, err = json.Marshal(data)
	} else {
		js, err = json.MarshalIndent(data, " ", "\t")
	}

	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *App) EncodeJson(data Envelope) ([]byte, error) {

	js, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	js = append(js, '\n')
	return js, err
}

func (app *App) ReadJSON(w http.ResponseWriter, r *http.Request, destination interface{}) error {

	err := json.NewDecoder(r.Body).Decode(destination)
	if err != nil {

		var SyntaxError *json.SyntaxError
		var UnmarshalTypeError *json.UnmarshalTypeError
		var InvalidUnmarshallError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &SyntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", SyntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &UnmarshalTypeError):
			if UnmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", UnmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", UnmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.As(err, &InvalidUnmarshallError):
			panic(err)

		default:
			return err
		}
	}
	return nil
}
