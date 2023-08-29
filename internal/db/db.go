package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// This is a file db, it just creates a json if it doesnt exist and writes given models on it,
type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"Id"`
}
type User struct {
	Email    string `json:"email"`
	Id       int    `json:"Id"`
	Password string `json:"password"`
}
type UserView struct {
	Email string `json:"email"`
	Id    int    `json:"Id"`
}
type UserLogin struct {
	User
	Expires_in_seconds int `json:"expires_in_seconds"`
}
type DB struct {
	path string
	mu   *sync.RWMutex
}
type DBStructure struct {
	Chirps       map[int]Chirp
	Users        map[int]User
	RefreshToken map[string]RefreshToken
}
type RefreshToken struct {
	Id         string
	Revoked    bool
	RevokeTime time.Time
}

func NewDB(path string) (*DB, error) {
	db := DB{path: path, mu: &sync.RWMutex{}}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}

	return &db, err
}
func (db *DB) CreateRefreshToken(refreshStr string, expiration time.Time) (RefreshToken, error) {
	db.ensureDB()
	str, err := db.loadDB()
	refresh := RefreshToken{}
	if err != nil {
		return refresh, err
	}
	refresh.Id = refreshStr
	refresh.Revoked = false
	refresh.RevokeTime = expiration

	str.RefreshToken[refreshStr] = refresh
	db.writeDB(str)
	return refresh, nil
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
func validateEmail(email string, str DBStructure) bool {
	users := str.Users
	for _, val := range users {
		if val.Email == email {
			return false
		}
	}
	return true
}
func (db *DB) CreateUser(email, password string) (UserView, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return UserView{}, err
	}
	emailOk := validateEmail(email, str)
	if !emailOk {
		return UserView{}, errors.New("Email already registered")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	user := User{Email: email, Password: string(hashed)}

	nextId := len(str.Users) + 1
	user.Id = nextId
	str.Users[nextId] = user

	err = db.writeDB(str)
	if err != nil {

		return UserView{}, err
	}

	return UserView{Email: user.Email, Id: user.Id}, nil

}
func (db *DB) GetUserByEmail(email string) (*User, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return &User{}, err
	}
	for _, value := range str.Users {
		if value.Email == email {
			return &value, nil
		}
	}
	return &User{}, errors.New("User not found")
}
func (db *DB) UpdateUser(old User, new User) (UserView, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return UserView{}, err
	}
	emailOk := validateEmail(new.Email, str)
	if !emailOk {
		return UserView{}, errors.New("This email already registered")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(new.Password), 4)
	user := User{}
	for i, userA := range str.Users {
		if userA.Email == old.Email {
			new.Password = string(hashed)
			new.Id = old.Id
			userA = new

			str.Users[i] = userA
			user = userA
			break
		}
	}

	db.writeDB(str)
	return UserView{
		Id:    user.Id,
		Email: user.Email,
	}, nil

}
func (db *DB) LoginUser(userLogin UserLogin) (UserView, error) {
	email := userLogin.Email
	password := userLogin.Password
	fmt.Println(userLogin.Expires_in_seconds)
	db.ensureDB()
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return UserView{}, err
	}
	compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if compareErr != nil {
		return UserView{}, errors.New("Login failed")
	}
	return UserView{Email: user.Email, Id: user.Id}, nil

}
func (db *DB) GetUserById(id int) (UserView, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return UserView{}, err
	}
	user, ok := str.Users[id]
	if !ok {
		return UserView{}, errors.New("User cannot found")
	}
	return UserView{user.Email, user.Id}, nil
}
func (db *DB) GetUserByIdORIGINAL(id int) (User, error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, ok := str.Users[id]
	if !ok {
		return User{}, errors.New("User cannot found")
	}
	return user, nil
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
