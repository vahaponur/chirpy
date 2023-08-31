package db

import (
	"errors"
	"fmt"
	"sort"
)

type Chirp struct {
	Body      string `json:"body"`
	Id        int    `json:"id"`
	Author_Id int    `json:"author_id"`
}

func (db *DB) CreateChirp(newChirp Chirp) (Chirp, error) {
	db.ensureDB()
	str, err := db.loadDB()
	chirp := Chirp{Body: newChirp.Body, Author_Id: newChirp.Author_Id}
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
func (db *DB) GetChirpValues(sortOpt string) ([]Chirp, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(str.Chirps))
	for _, chirp := range str.Chirps {
		chirps = append(chirps, chirp)
	}
	if sortOpt != "desc" {
		sortByAsc(chirps)
	} else {
		sortByDesc(chirps)
	}

	return chirps, nil
}
func (db *DB) GetChirpById(id int) (Chirp, error) {
	db.ensureDB()

	str, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := str.Chirps[id]
	if !ok {
		return Chirp{}, errors.New(fmt.Sprintf("Could not find a chirp with ID: %v", id))
	}
	return chirp, nil

}
func (db *DB) DeleteChirpById(id int) error {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return err
	}
	chirp, err := db.GetChirpById(id)
	if err != nil {
		return err
	}

	delete(str.Chirps, chirp.Id)
	db.writeDB(str)
	return nil
}
func (db *DB) GetAuthorChirps(authorId int, sortOpt string) ([]Chirp, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := make([]Chirp, 0, 0)
	for _, chirp := range str.Chirps {
		if chirp.Author_Id == authorId {
			chirps = append(chirps, chirp)
		}
	}
	if sortOpt != "desc" {
		sortByAsc(chirps)
	} else {
		sortByDesc(chirps)
	}

	return chirps, nil
}
func sortByAsc(chirps []Chirp) {
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})
}
func sortByDesc(chirps []Chirp) {
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id > chirps[j].Id
	})
}
