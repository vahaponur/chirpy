package main

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"internal/db"
	"io"
	"net/http"
	"strconv"
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
	tokenStr, err := getAccessToken(param)
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

func updateUser(res http.ResponseWriter, req *http.Request) {
	tokenString, err := getCleanTokenStr(req)
	if err != nil {
		respondWithError(res, http.StatusBadRequest, err.Error())
		return
	}
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
