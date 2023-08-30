package db

import (
	"errors"
	"time"
)

type RefreshToken struct {
	Id         string
	Revoked    bool
	RevokeTime time.Time
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
func (db *DB) GetRefreshToken(refreshStr string) (rt RefreshToken, err error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return
	}
	rt, ok := str.RefreshToken[refreshStr]
	if !ok {
		err = errors.New("Refresh token is invalid")
		return
	}

	return
}
func (db *DB) RevokeRefreshToken(token RefreshToken) (rt RefreshToken, err error) {
	db.ensureDB()
	str, err := db.loadDB()
	if err != nil {
		return
	}
	rt, ok := str.RefreshToken[token.Id]
	if !ok {
		err = errors.New("Refresh token is invalid")
		return
	}
	rt.Revoked = true
	rt.RevokeTime = time.Now()
	str.RefreshToken[token.Id] = rt
	db.writeDB(str)
	return db.GetRefreshToken(rt.Id)
}
