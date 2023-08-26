package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileServerHit int
}

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

func main() {
	cfg := apiConfig{fileServerHit: 0}

	apiRooter := chi.NewRouter()
	apiRooter.Get("/healthz", healthzHandler)
	apiRooter.Post("/validate_chirp", validationHandler)
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

func validationHandler(res http.ResponseWriter, req *http.Request) {
	type params struct {
		Body string `json:"body"`
	}
	param := params{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	json.Unmarshal(body, &param)

	if len(param.Body) > 140 {
		respondWithError(res, http.StatusBadRequest, "Chirp is too long")
		return
	}
	current := param.Body
	for _, fk := range FORBIDDEN_KEYWORDS {
		if strings.Contains(strings.ToLower(current), fmt.Sprintf(" %v ", fk)) {
			current = strings.Replace(current, fk, "****", -1)
			current = strings.Replace(current, strings.Title(fk), "****", -1)
		}
	}
	res.Header().Set("Content-Type", "application/json")
	type MyData struct {
		Data string `json:"cleaned_body"`
	}
	respondWithJSON(res, 200, MyData{Data: current})

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
