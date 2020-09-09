package main

import (
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	//Router http.Handler
	repo   *Repo
	Router *chi.Mux
	build  string
	//authenticator *auth.Authenticator
}

type Service struct {
	repo *Repo
}

// NewServer is a factory function which creates and initializes new user REST API server.
func NewServer(build string, dbName string, mngClient *mongo.Client) *Server {
	s := Server{
		build: build,
		repo:  NewRepo(mngClient, dbName),
	}

	s.initRoutes()
	return &s
}

