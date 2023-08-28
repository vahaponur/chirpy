package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

var FORBIDDEN_KEYWORDS = []string{"kerfuffle", "sharbert", "fornax"}

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
