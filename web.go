package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	headerContentType   = "Content-Type"
	mimeApplicationJSON = "application/json"
)

func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(ioutil.Discard, r)
	return json.NewDecoder(r).Decode(v)
}

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", headerContentType, "X-CSRF-Token"},
		ExposedHeaders: []string{"X-Total-Count"},
	})
	r.Use(corsMiddleware.Handler)
	return r
}

// RespondError create json error response and outputs passed error into response body.
func RespondError(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	Respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...)},
	})
}

// Respond create json response and outputs json representation of the passed data into response body.
func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set(headerContentType, mimeApplicationJSON)
	w.WriteHeader(status)
	if data != nil {
		err := EncodeBody(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// EncodeBody encodes passed date to json format and writes it into Response body.
func EncodeBody(w http.ResponseWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func URLParamObjectID(r *http.Request, key string) (primitive.ObjectID, error) {
	profileID := chi.URLParam(r, key)
	if profileID == "" {
		return primitive.NilObjectID, errors.New("invalid request data " + key)
	}

	ObjID, err := primitive.ObjectIDFromHex(profileID)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid request data " + key)
	}

	return ObjID, nil
}
