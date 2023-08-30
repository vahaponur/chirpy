package db

import (
	"encoding/json"
	"os"
	"sync"
)

// This is a file db, it just creates a json if it does not exist and writes given models on it,

type DB struct {
	path string
	mu   *sync.RWMutex
}
type DBStructure struct {
	Chirps       map[int]Chirp
	Users        map[int]User
	RefreshToken map[string]RefreshToken
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
	dbStructure := DBStructure{Chirps: map[int]Chirp{}, Users: map[int]User{}, RefreshToken: map[string]RefreshToken{}}
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
