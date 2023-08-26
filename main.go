package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	type returnvals struct {
		Valid bool `json:"valid"`
	}
	type errorvals struct {
		Error string `json:"error"`
	}
	if len(param.Body) > 140 {
		res.WriteHeader(http.StatusBadRequest)
		tooLongErr := errorvals{Error: "Chirp is too long"}
		data, err := json.Marshal(tooLongErr)
		if err != nil {

			res.Write([]byte(err.Error()))
		}
		res.Write(data)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	data, err := json.Marshal(returnvals{Valid: true})
	res.Write(data)

}
