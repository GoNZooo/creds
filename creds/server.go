package creds

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	router *httprouter.Router
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.router.ServeHTTP(writer, request)
}

type DatabaseOptions struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func (server *Server) Serve(port int, database *pg.DB, adminScope string) {
	if server.router == nil {
		server.router = httprouter.New()
	}

	setupRoutes(server.router, database, adminScope)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), server); err != nil {
		log.Panicf("Error trying to start server: %server", err.Error())
	}
}
