package creds

import (
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/julienschmidt/httprouter"
)

func setupRoutes(router *httprouter.Router, database *pg.DB, adminScope string) {
	routes := []routeSpecification{
		post{"/tokens", handleAddToken(database, adminScope)},
		post{"/users", handleAddUser(database, adminScope)},
		get{"/users", handleGetUsers(database, adminScope)},
		get{"/user/:Id", handleGetUser(database, adminScope)},
	}

	addRoutes(router, routes)
}

type post struct {
	path    string
	handler http.HandlerFunc
}

func (post post) toRouteData() routeData {
	return routeData{
		method:  "POST",
		path:    post.path,
		handler: post.handler,
	}
}

type get struct {
	path    string
	handler http.HandlerFunc
}

func (get get) toRouteData() routeData {
	return routeData{
		method:  "GET",
		path:    get.path,
		handler: get.handler,
	}
}

type routeData struct {
	method  string
	path    string
	handler http.HandlerFunc
}

type routeSpecification interface {
	toRouteData() routeData
}

func addRoutes(router *httprouter.Router, routes []routeSpecification) {
	for _, r := range routes {
		rd := r.toRouteData()
		addRouteData(router, rd)
	}
}

func addRouteData(router *httprouter.Router, rd routeData) {
	router.HandlerFunc(rd.method, rd.path, rd.handler)
}
