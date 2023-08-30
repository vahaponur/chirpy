package main

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"internal/db"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
func getAccessToken(userLogin db.UserLogin) (string, error) {
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
func getCleanTokenStr(req *http.Request) (string, error) {
	bearer := req.Header.Get("Authorization")
	tokenString, ok := strings.CutPrefix(bearer, "Bearer ")
	if !ok {
		return "", errors.New("Bearer could not stripped")
	}
	return tokenString, nil
}
func updateAccessToken(res http.ResponseWriter, req *http.Request) {
	token, err := getCleanTokenStr(req)
	//Test caseler 401 istiyo burasi bad request olmali
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	claims, err := getTokenClaims(token)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	if claims.Issuer != "chirpy-refresh" {
		respondWithError(res, http.StatusUnauthorized, "Cannot get a token without refresh token")
		return
	}
	refreshToken, err := Db.GetRefreshToken(token)
	//Burasi internal server error veya bad request olmali
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	if refreshToken.Revoked {
		respondWithError(res, http.StatusUnauthorized, "Refresh token revoked from the system")
		return
	}
	userLogin := db.UserLogin{}
	userLogin.Id, err = strconv.Atoi(claims.Subject)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}
	newAccessToken, err := getAccessToken(userLogin)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}
	type Token struct {
		Token string `json:"token"`
	}
	respondWithJSON(res, http.StatusOK, Token{newAccessToken})
}
func revokeRefreshToken(res http.ResponseWriter, req *http.Request) {
	token, err := getCleanTokenStr(req)
	//Test caseler 401 istiyo burasi bad request olmali
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	claims, err := getTokenClaims(token)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	if claims.Issuer != "chirpy-refresh" {
		respondWithError(res, http.StatusUnauthorized, "Cannot get a token without refresh token")
		return
	}
	refreshToken, err := Db.GetRefreshToken(token)
	//Burasi internal server error veya bad request olmali
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	if refreshToken.Revoked {
		respondWithError(res, http.StatusUnauthorized, "Refresh token revoked from the system before")
		return
	}
	revoked, err := Db.RevokeRefreshToken(refreshToken)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	if revoked.Revoked {
		respondWithJSON(res, http.StatusOK, revoked)
		return
	}

}
