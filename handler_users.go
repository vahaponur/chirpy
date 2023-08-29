package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"internal/db"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func addUser(res http.ResponseWriter, req *http.Request) {

	param := db.User{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	json.Unmarshal(body, &param)

	res.Header().Set("Content-Type", "application/json")

	user, err := Db.CreateUser(param.Email, param.Password)
	if err != nil {
		respondWithError(res, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(res, 201, user)

}
func loginUser(res http.ResponseWriter, req *http.Request) {
	param := db.UserLogin{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	json.Unmarshal(body, &param)

	res.Header().Set("Content-Type", "application/json")

	if err != nil {
		respondWithError(res, http.StatusBadRequest, err.Error())
	}

	user, err := Db.LoginUser(param)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	param.Id = user.Id
	tokenStr, err := getJWT(param)
	login := Login{UserView: user, Token: tokenStr}
	respondWithJSON(res, 200, login)

}

type Login struct {
	UserView db.UserView
	Token    string `json:"token"`
}

func getJWT(userLogin db.UserLogin) (string, error) {
	numDate := time.Now()
	expiration := numDate.Add(24 * time.Hour)
	if userLogin.Expires_in_seconds != 0 {
		expiration = numDate.Add(time.Duration(userLogin.Expires_in_seconds) * time.Second)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   strconv.Itoa(userLogin.Id),
		Audience:  nil,
		ExpiresAt: &jwt.NumericDate{Time: expiration},
		NotBefore: nil,
		IssuedAt:  &jwt.NumericDate{Time: numDate},
		ID:        "",
	})
	fmt.Println(userLogin.Id)

	jwtS, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		return jwtS, err
	}
	return jwtS, nil

}

func updateUser(res http.ResponseWriter, req *http.Request) {
	bearer := req.Header.Get("Authorization")
	tokenString, _ := strings.CutPrefix(bearer, "Bearer ")
	fmt.Println(tokenString)
	type MyCustomClaims struct {
		Foo string `json:"foo"`
		jwt.RegisteredClaims
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil

	})
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	claims, ok := token.Claims.(*MyCustomClaims)
	if !ok {
		respondWithError(res, http.StatusUnauthorized, "Sbisioldu")
		return
	}
	if !token.Valid {
		respondWithError(res, http.StatusUnauthorized, "You shall not pass")
		return
	}
	fmt.Println(claims.Foo)
	fmt.Println(claims.IssuedAt)
	fmt.Println(claims.RegisteredClaims)
	userId := claims.Subject
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	fmt.Println(userId)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := Db.GetUserById(userIdInt)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	param := db.User{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	json.Unmarshal(body, &param)
	uw, err := Db.UpdateUser(user.Email, param.Email)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(res, 200, uw)
}
