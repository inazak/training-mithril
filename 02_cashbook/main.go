package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

//go:embed static
var embedFS embed.FS

func main() {

	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/api/cashbook", func(r chi.Router) {
		r.Post("/", handleCashbookEntry)
	})
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})
	http.ListenAndServe(":4989", r)
}

type cashbookEntry struct {
	Date   string `json:"date" validate:"required"`
	Item   string `json:"item" validate:"required"`
	Amount string `json:"amount" validate:"required"`
}

type cashbookEntries struct {
	Entries []cashbookEntry `json:"entries" validate:reqired,dive,required`
}

type errMessage struct {
	Message string `json:"message"`
}

func handleCashbookEntry(w http.ResponseWriter, r *http.Request) {
	validation := validator.New()

	var body cashbookEntries
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(errMessage{
			Message: "fail to decode json",
		}, http.StatusInternalServerError)(w, r)
		return
	}
	if err := validation.Struct(body); err != nil {
		respondJSON(errMessage{
			Message: "fail to validate json " + err.Error(),
		}, http.StatusBadRequest)(w, r)
		return
	}

	fmt.Printf("[debug] body=%#v\n", body)
	respondJSON(body, http.StatusCreated)(w, r)
}

func respond(body string, ctype string, status int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ctype+"; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, "%s", body)
	}
}

func respondJSON(body any, status int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		respond(string(b), "text/json", status)(w, r)
	}
}
