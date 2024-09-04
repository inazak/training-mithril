package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/yuin/goldmark"
	"go.abhg.dev/goldmark/wikilink"
	"io/fs"
	"net/http"
)

var validation = validator.New()

var wikipage = map[string]string{
	"home":  "姫路城（ひめじじょう）は、兵庫県姫路市にある日本の城。[[江戸時代]]初期に建てられた天守や櫓等の主要建築物が現存し、国宝や重要文化財に指定されている。[[ユネスコ]]の世界遺産（文化遺産）リストにも登録され、日本100名城などに選定されている。",
	"江戸時代": "江戸時代（えどじだい）は、日本の歴史の内江戸幕府（徳川幕府）の統治時代を指す時代区分である。他の呼称として徳川時代、徳川日本、旧幕時代、藩政時代（藩領のみ）などがある。江戸時代という名は、江戸に将軍が常駐していたためである。 ",
	"ユネスコ": "国際連合教育科学文化機関（英: United Nations Educational, 略称: UNESCO、ユネスコ）は、国際連合の経済社会理事会の下におかれた、教育、科学、文化の発展と推進、世界遺産の登録などを目的とした国際協定である。",
}

type responsePage struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	Raw    string `json:"raw"`
	HTML   string `json:"html"`
}

type responsePageList struct {
	Status string   `json:"status"`
	IdList []string `json:"idlist"`
}

type responseError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type requestPage struct {
	Raw string `json:"raw"`
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
	r.Route("/api/wiki/page", func(r chi.Router) {
		r.Get("/", handleGetPageList)
		r.Get("/{pageID}", handleGetPage)
		r.Post("/{pageID}", handlePostPage)
	})
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})
	http.ListenAndServe(":4989", r)
}

func handleGetPageList(w http.ResponseWriter, r *http.Request) {
	idlist := []string{}
	for id, _ := range wikipage {
		idlist = append(idlist, id)
	}

	respondJSON(responsePageList{
		Status: "ok",
		IdList: idlist,
	}, http.StatusOK)(w, r)
}

func handleGetPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "pageID")
	body, ok := wikipage[id]

	if !ok {
		respondJSON(responseError{
			Status:  "NG",
			Message: "notfound",
		}, http.StatusNotFound)(w, r)
		return
	}

	html, _ := convert(body)
	respondJSON(responsePage{
		Status: "OK",
		Id:     id,
		Raw:    body,
		HTML:   html,
	}, http.StatusOK)(w, r)
}

func handlePostPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "pageID")
	var body requestPage

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(responseError{
			Status:  "NG",
			Message: "fail to json decode",
		}, http.StatusInternalServerError)(w, r)
		return
	}

	if err := validation.Struct(body); err != nil {
		respondJSON(responseError{
			Status:  "NG",
			Message: "fail to json validation",
		}, http.StatusBadRequest)(w, r)
		return
	}

	wikipage[id] = body.Raw

	respondJSON(responsePage{
		Status: "OK",
		Id:     id,
		Raw:    "",
		HTML:   "",
	}, http.StatusCreated)(w, r)
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

type addPrefixResolver struct {
	prefix string
}

func (r addPrefixResolver) ResolveWikilink(node *wikilink.Node) ([]byte, error) {
	return []byte(r.prefix + string(node.Target)), nil
}

func convert(source string) (string, error) {
	md := goldmark.New(goldmark.WithExtensions(&wikilink.Extender{
		Resolver: addPrefixResolver{prefix: "#!/page/"},
	}))
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
