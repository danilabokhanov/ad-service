package dbrepo

import "adservice/internal/app"

func New(db DataBase) app.Repository {
	return &DBRepo{db: db}
}
