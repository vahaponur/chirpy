package main

import (
	"encoding/json"
	"internal/db"
	"io"
	"net/http"
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
	param := db.User{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("Something went wrong"))
		return
	}
	json.Unmarshal(body, &param)

	res.Header().Set("Content-Type", "application/json")
	user, err := Db.LoginUser(param.Email, param.Password)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, err.Error())
		return
	}
	respondWithJSON(res, 200, user)

}
