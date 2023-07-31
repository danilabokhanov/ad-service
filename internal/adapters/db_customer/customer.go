package dbcustomer

import (
	"adservice/internal/app"
)

func New(db DataBase) app.Users {
	return &DBCustomer{db: db}
}
