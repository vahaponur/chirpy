package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"internal/db"

	"log"
	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileServerHit int
	jwtSecret     string
}

var Db *db.DB
var DbStructure *db.DBStructure
var cfg *apiConfig

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

var debugMode *bool

func getMainRouter() *chi.Mux {
	cfg = &apiConfig{fileServerHit: 0, jwtSecret: os.Getenv("JWT_SECRET")}
	apiRooter := chi.NewRouter()
	apiRooter.Get("/healthz", healthzHandler)
	apiRooter.Post("/chirps", addChirp)
	apiRooter.Get("/chirps", getChirps)
	apiRooter.Get("/chirps/{id}", getChirpById)
	apiRooter.Delete("/chirps/{id}", deleteChirpById)
	apiRooter.Post("/users", addUser)
	apiRooter.Post("/login", loginUser)
	apiRooter.Put("/users", updateUser)
	apiRooter.Post("/refresh", updateAccessToken)
	apiRooter.Post("/revoke", revokeRefreshToken)

	adminRooter := chi.NewRouter()
	adminRooter.Get("/metrics", cfg.metrics)
	r := chi.NewRouter()
	fsHandler := cfg.incMetrics(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)
	r.Mount("/api", apiRooter)
	r.Mount("/admin", adminRooter)
	r.Handle("/assets", http.FileServer(http.Dir("./assets")))
	return r
}
func main() {
	debugMode = flag.Bool("debug", false, "Enable debug mode")
	godotenv.Load()
	defer cleanup()

	CreateDb()

	r := getMainRouter()
	corsmux := middlewareCors(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsmux,
	}
	fmt.Println("Starting the server...")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error: %s\n", err)
		}
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	// Wait for signals
	<-signalChan
	fmt.Println("\nReceived interrupt signal. Shutting down gracefully...")

	// Shutdown the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Error during server shutdown: %s\n", err)
	}

}
func cleanup() {
	flag.Parse()
	if !*debugMode {
		return
	}
	err := os.Remove("./database.json")
	if err != nil {
		fmt.Println(err)
	}
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

func respondWithError(w http.ResponseWriter, code int, msg interface{}) {
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
