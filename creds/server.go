package creds

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/julienschmidt/httprouter"
)

var _ = pg.DB{}
var _ = http.Client{}
var _ = httprouter.New()

type Server struct {
	router *httprouter.Router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

type DatabaseOptions struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func (s *Server) Serve(port int, databaseOptions DatabaseOptions, adminScope string) {
	if s.router == nil {
		s.router = httprouter.New()
	}

	db := connectToDatabase(databaseOptions)
	err := createSchema(db)
	if err != nil {
		log.Panicf("`createSchema` error: %s", err.Error())
	}

	s.setupRoutes(db, adminScope)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), s); err != nil {
		log.Panicf("Error trying to start server: %s", err.Error())
	}
}
