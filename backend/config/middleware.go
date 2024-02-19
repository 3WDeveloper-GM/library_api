package config

import (
	"net/http"
	"net/http/httptest"
)

func (app *App) VisitedRouteLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var message = "got a request with the following:"

		app.Log.Info().Interface("request data", struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}{
			Method: r.Method,
			Path:   r.URL.Path,
		}).
			Msg(message)

		c := httptest.NewRecorder()
		next.ServeHTTP(c, r)

		for k, v := range c.Result().Header {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		c.Body.WriteTo(w)

		message = "sent the following response:"

		if res := c.Result().StatusCode; res <= 300 {
			app.Log.Info().Interface("response data", struct {
				Status  string      `json:"status"`
				Headers interface{} `json:"headers"`
			}{
				Status:  c.Result().Status,
				Headers: c.Result().Header,
			}).
				Msg(message)
		} else {
			app.Log.Error().Interface("response data", struct {
				StatusCode string      `json:"status"`
				Headers    interface{} `json:"headers"`
			}{
				StatusCode: c.Result().Status,
				Headers:    c.Result().Header,
			}).
				Msg(message)
		}

	})
}
