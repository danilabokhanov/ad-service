package dbcustomer

import (
	"adservice/internal/user"
	"context"
)

type DBCustomer struct {
	db DataBase
}

type DataBase interface {
	SelectById(ctx context.Context, userID int64) ([]user.User, error)
	UpdateById(ctx context.Context, userID int64, nickname, email string) error
	Insert(ctx context.Context, nickname string, email string, userID int64) error
	DeleteById(ctx context.Context, userID int64) error
	Close() error
}

func (d *DBCustomer) Find(ctx context.Context, userID int64) (user.User, bool) {
	u, err := d.db.SelectById(ctx, userID)
	if err != nil {
		return user.User{}, false
	}
	if len(u) != 1 {
		return user.User{}, false
	}
	return u[0], true
}

func (d *DBCustomer) ChangeInfo(ctx context.Context, userID int64, nickname, email string) error {
	return d.db.UpdateById(ctx, userID, nickname, email)
}

func (d *DBCustomer) CreateByID(ctx context.Context, nickname string, email string, userID int64) (user.User, error) {
	return user.User{Nickname: nickname, Email: email, ID: userID}, d.db.Insert(ctx, nickname, email, userID)
}

func (d *DBCustomer) DeleteByID(ctx context.Context, userID int64) error {
	return d.db.DeleteById(ctx, userID)
}
