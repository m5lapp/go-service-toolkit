package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/m5lapp/go-service-toolkit/validator"
)

func (app *WebApp) ReadIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *WebApp) ReadString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *WebApp) ReadCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *WebApp) ReadInt(qs url.Values, key string, defaultValue int,
	v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *WebApp) Background(fn func()) {
	app.Wg.Add(1)
	go func() {
		defer app.Wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				// As recover() returns an any type, create an error out of it.
				e := fmt.Errorf("%s", err)
				app.Logger.Error(e.Error(), nil)
			}
		}()

		fn()
	}()
}
