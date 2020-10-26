package creds

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Tries to get the `httprouter.Param`s from a request. If it's unable to, will return `nil`
func getParameters(request *http.Request) httprouter.Params {
	parameters, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)

	if ok {
		return parameters
	} else {
		return nil
	}
}
