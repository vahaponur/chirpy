package db

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
)

type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"Id"`
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.ensureDB()
	str, err := db.loadDB()
	chirp := Chirp{Body: body}
	if err != nil {
		return chirp, err
	}
	nextId := str.Chirps[len(str.Chirps)].Id + 1
	chirp.Id = nextId
	str.Chirps[nextId] = chirp
	err = db.writeDB(str)
	if err != nil {
		return chirp, err
	}
	return chirp, err

}

// GetChirps returns all chirps in the database
func (db *DB) GetChirpValues() ([]Chirp, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirpValues := make([]Chirp, 0, len(str.Chirps))
	for _, chirp := range str.Chirps {
		chirpValues = append(chirpValues, chirp)
	}
	sort.Slice(chirpValues, func(i, j int) bool { return chirpValues[i].Id < chirpValues[j].Id })
	return chirpValues, nil
}
func (db *DB) GetChirpById(id string) (Chirp, error) {
	db.ensureDB()
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return Chirp{}, err
	}
	str, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := str.Chirps[idInt]
	if !ok {
		return Chirp{}, errors.New(fmt.Sprintf("Could not find a chirp with ID: %v", idInt))
	}
	return chirp, nil

}