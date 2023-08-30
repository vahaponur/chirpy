package main

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"internal/db"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type CustomClaims struct {
	jwt.RegisteredClaims
}
type Login struct {
	db.UserView
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
type PolkaRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

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
func upgradeUser(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	apikey := req.Header.Get("Authorization")
	apiCutted, ok := strings.CutPrefix(apikey, "ApiKey ")
	if !ok {
		respondWithError(res, http.StatusUnauthorized, "")
		return
	}
	if apiCutted != cfg.polkaKey {
		respondWithError(res, http.StatusUnauthorized, "")
		return
	}
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}
	polka := PolkaRequest{}
	json.Unmarshal(body, &polka)
	if polka.Event != "user.upgraded" {
		res.WriteHeader(200)
		return
	}
	err = Db.UpgradeUser(polka.Data.UserID)
	if err != nil {
		respondWithError(res, http.StatusNotFound, err.Error())
		return
	}
	respondWithJSON(res, http.StatusOK, "")

}
