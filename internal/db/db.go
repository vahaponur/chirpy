package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
)

// This is a file db, it just creates a json if it doesnt exist and writes given models on it,
type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"id"`
}
type DB struct {
	path string
	mu   *sync.RWMutex
}
type DBStructure struct {
	Chirps map[int]Chirp
}

func NewDB(path string) (*DB, error) {
	db := DB{path: path, mu: &sync.RWMutex{}}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}

	return &db, err
}

// CreateChirp creates a new chirp and saves it to disk
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

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, err := os.Stat(db.path)

	if os.IsNotExist(err) {
		file, err := os.Create(db.path)
		defer file.Close()
		if err != nil {
			return err
		}

	} else {
		return err
	}
	return nil

}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	dbStructure := DBStructure{Chirps: map[int]Chirp{}}
	file, err := os.ReadFile(db.path)
	if err != nil {
		return dbStructure, err
	}

	if len(file) == 0 {
		// If the file is empty, unmarshal the default empty JSON structure
		return dbStructure, nil
	}
	err = json.Unmarshal(file, &dbStructure)
	if err != nil {

		return dbStructure, err
	}
	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}
