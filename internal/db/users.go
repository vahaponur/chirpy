package db

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email       string `json:"email"`
	Id          int    `json:"id"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}
type UserView struct {
	Email       string `json:"email"`
	Id          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}
type UserLogin struct {
	User
	Expires_in_seconds int `json:"expires_in_seconds"`
}

func UserToView(user User) (view UserView) {
	view.Id = user.Id
	view.IsChirpyRed = user.IsChirpyRed
	view.Email = user.Email
	return
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

	return UserToView(user), nil

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
	return UserToView(user), nil

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
	return UserToView(*user), nil

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
	return UserToView(user), nil
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
func (db *DB) UpgradeUser(userId int) error {

	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return err
	}
	user, ok := str.Users[userId]
	if !ok {
		return errors.New("User could not found")
	}
	user.IsChirpyRed = true
	str.Users[userId] = user
	db.writeDB(str)
	return nil
}
