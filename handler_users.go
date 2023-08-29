package main

import (
	"encoding/json"
	"errors"
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
		respondWithError(res, http.StatusBadRequest, err.Error())
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

	user, err := Db.LoginUser(param)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	param.Id = user.Id
	tokenStr, err := getToken(param)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
	}
	refreshStr, err := getRefreshToken(param)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
	}

	login := Login{UserView: user, Token: tokenStr, RefreshToken: refreshStr}
	respondWithJSON(res, 200, login)

}

type Login struct {
	db.UserView
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func getToken(userLogin db.UserLogin) (string, error) {
	numDate := time.Now()
	expiration := numDate.Add(time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
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
func getRefreshToken(userLogin db.UserLogin) (string, error) {
	numDate := time.Now()
	expiration := numDate.Add(24 * 60 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
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
		return "", err
	}
	//Refresh Tokens saved to DB
	refreshTokenStr, err := Db.CreateRefreshToken(jwtS, expiration)
	if err != nil {
		return "", err
	}
	return refreshTokenStr.Id, nil

}
func updateUser(res http.ResponseWriter, req *http.Request) {
	bearer := req.Header.Get("Authorization")
	tokenString, _ := strings.CutPrefix(bearer, "Bearer ")

	claims, err := getTokenClaims(tokenString)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	if claims.Issuer == "chirpy-refresh" {
		respondWithError(res, http.StatusUnauthorized, "Cannot access with a refresh token")
		return
	}
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
	user, err := Db.GetUserByIdORIGINAL(userIdInt)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}
	param := db.User{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}
	json.Unmarshal(body, &param)

	uw, err := Db.UpdateUser(user, param)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(res, 200, uw)
}

type CustomClaims struct {
	jwt.RegisteredClaims
}

func getTokenClaims(tokenString string) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil

	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {

		return nil, errors.New("Claims cannot be redeemed")
	}
	if !token.Valid {

		return nil, errors.New("Token is invalid")
	}
	return claims, nil
}
