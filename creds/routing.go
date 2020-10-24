package creds

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v10"
)

func (s *Server) setupRoutes(_ *pg.DB) {
	routes := []routeSpecification{
		get{"/", handleIndex()},
	}

	s.addRoutes(routes)
}

func handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "hello!")
		if err != nil {
			fmt.Printf("Unable to send hello response: %s", err.Error())
		}
	}
}

type post struct {
	path    string
	handler http.HandlerFunc
}

func (p post) toRouteData() routeData {
	return routeData{
		method:  "POST",
		path:    p.path,
		handler: p.handler,
	}
}

type get struct {
	path    string
	handler http.HandlerFunc
}

func (g get) toRouteData() routeData {
	return routeData{
		method:  "GET",
		path:    g.path,
		handler: g.handler,
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

func (s *Server) addRoutes(routes []routeSpecification) {
	for _, r := range routes {
		rd := r.toRouteData()
		s.addRouteData(rd)
	}
}

func (s *Server) addRouteData(rd routeData) {
	s.router.HandlerFunc(rd.method, rd.path, rd.handler)
}

var _ = post{}
var _ = get{}