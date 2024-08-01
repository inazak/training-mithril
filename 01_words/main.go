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

type Word struct {
	Id   int    `json:"id"`
	Word string `json:"word"`
}

var WordList = []Word{
	{Id: 0, Word: "red"},
	{Id: 1, Word: "blue"},
	{Id: 2, Word: "yellow"},
}

//go:embed static
var embedFS embed.FS

func main() {

	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/api/word", func(r chi.Router) {
		r.Get("/", handleGetWordList)
		r.Post("/", handlePostWord)
	})
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})
	http.ListenAndServe(":4989", r)
}

func handleGetWordList(w http.ResponseWriter, r *http.Request) {
	respondJSON(WordList, http.StatusOK)(w, r)
}

func handlePostWord(w http.ResponseWriter, r *http.Request) {
	validation := validator.New()
	var body struct {
		Word string `json:"word" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := validation.Struct(body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := len(WordList)
	WordList = append(WordList, Word{Id: id, Word: body.Word})
	respondJSON(WordList[id], http.StatusCreated)(w, r)
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
