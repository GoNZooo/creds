package creds

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Tries to get the `httprouter.Param`s from a request. If it's unable to, will return `nil`
func getParameters(r *http.Request) httprouter.Params {
	parameters, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)

	if ok {
		return parameters
	} else {
		return nil
	}
}
