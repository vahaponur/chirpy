package main

import (
	"encoding/json"
	"fmt"
	"internal/db"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileServerHit int
}

var Db *db.DB
var DbStructure *db.DBStructure

func (cfg *apiConfig) incMetrics(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileServerHit++
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(fmt.Sprintf(`<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>`, cfg.fileServerHit)))

}
func CreateDb() {
	Dba, err := db.NewDB("./database.json")
	if err != nil {
		log.Fatal(err)
	}
	Db = Dba
}
func main() {
	cfg := apiConfig{fileServerHit: 0}
	CreateDb()
	apiRooter := chi.NewRouter()
	apiRooter.Get("/healthz", healthzHandler)
	apiRooter.Post("/chirps", addChirp)
	apiRooter.Get("/chirps", getChirps)
	apiRooter.Get("/chirps/{id}", getChirpById)
	adminRooter := chi.NewRouter()
	adminRooter.Get("/metrics", cfg.metrics)
	r := chi.NewRouter()
	fsHandler := cfg.incMetrics(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)
	r.Mount("/api", apiRooter)
	r.Mount("/admin", adminRooter)
	r.Handle("/assets", http.FileServer(http.Dir("./assets")))

	corsmux := middlewareCors(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsmux,
	}
	server.ListenAndServe()

}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))

}

var FORBIDDEN_KEYWORDS = []string{"kerfuffle", "sharbert", "fornax"}

func addChirp(res http.ResponseWriter, req *http.Request) {

	param := Chirp{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	json.Unmarshal(body, &param)

	validation := validateChirp(req, param)
	if !validation.valid {
		respondWithError(res, http.StatusBadRequest, validation.message)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	type MyData struct {
		Data string `json:"cleaned_body"`
	}
	chirp, err := Db.CreateChirp(param.Body)
	if err != nil {
		respondWithError(res, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(res, 201, chirp)

}
func getChirps(res http.ResponseWriter, req *http.Request) {
	chirps, err := Db.GetChirpValues()
	if err != nil {
		respondWithError(res, 400, err.Error())
		return
	}
	respondWithJSON(res, 200, chirps)
}
func getChirpById(res http.ResponseWriter, req *http.Request) {
	param := chi.URLParam(req, "id")
	chirp, err := Db.GetChirpById(param)
	if err != nil {
		respondWithError(res, 404, err.Error())
		return
	}
	respondWithJSON(res, 200, chirp)
}

type Validation struct {
	valid   bool
	message string
}
type Chirp struct {
	Body string `json:"body"`
}

func validateChirp(req *http.Request, chirp Chirp) Validation {

	if len(chirp.Body) > 140 {

		return Validation{false, "Chirp is too long"}
	}
	current := chirp.Body
	for _, fk := range FORBIDDEN_KEYWORDS {
		if strings.Contains(strings.ToLower(current), fmt.Sprintf(" %v ", fk)) {
			current = strings.Replace(current, fk, "****", -1)
			current = strings.Replace(current, strings.Title(fk), "****", -1)
		}
	}
	return Validation{true, current}
}
func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	data, err := json.Marshal(msg)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(data)
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	w.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "error parsing payload")
	}
	w.Write(data)

}
