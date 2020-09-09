package main

import "github.com/go-chi/chi"

func (s *Server) initRoutes() {
	if s.Router == nil {
		summary := NewRouter()
		//auth := *s.authenticator

		// /api/v1/matching/
		summary.Group(func(r chi.Router) {
			//r.Use(web.Verifier(auth.JWTAuth()))
			//r.Use(web.Authenticator)
			r.Get("/", s.getMatchingsHandler)
			r.Get("/summary/{summaryId}", s.getMatchingHandler)
			r.Post("/", s.postMatchingHandler)
			r.Post("/bulk/", s.postMatchingsBulkHandler)
		})
		s.Router = summary
	}
}
