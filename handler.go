package main

import (
	"errors"
	"net/http"
)

// getMatchingsHandler is a handler function to return list of matchings
// endpoint: GET /api/v1/matching
func (s *Server) getMatchingsHandler(w http.ResponseWriter, r *http.Request) {
	tagGroups, err := s.repo.GetAllMatchings(r.Context())
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	Respond(w, r, http.StatusOK, tagGroups)
}

// getMatchingsHandler is a handler function to return list of matchings
// endpoint: GET /api/v1/matching/summary/{summaryId}
func (s *Server) getMatchingHandler(w http.ResponseWriter, r *http.Request) {
	summaryID, err := URLParamObjectID(r, "summaryId")
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, err)
		return
	}

	matching, err := s.repo.GetMatchingBySummaryId(r.Context(), summaryID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	Respond(w, r, http.StatusOK, matching)
}

// postMatchingsBulkHandler save new
// endpoint: POST /api/v1/matching
// payload:
func (s *Server) postMatchingHandler(w http.ResponseWriter, r *http.Request) {
	matching := Matching{}
	if err := DecodeJSON(r.Body, &matching); err != nil {
		RespondError(w, r, http.StatusBadRequest, err)
		return
	}

	modCount, err := s.repo.UpdateMatching(r.Context(), matching)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	if modCount == 0 {
		err := errors.New("Nothing was changed.")
		RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	Respond(w, r, http.StatusOK, matching)
}

// postMatchingsBulkHandler save new
// endpoint: POST /api/v1/matching/bulk
// payload:
func (s *Server) postMatchingsBulkHandler(w http.ResponseWriter, r *http.Request) {
	//decodeRequest()
}
