package basiccustomer

import (
	"adservice/internal/user"
	"context"
	"sync"
)

type BasicCustomer struct {
	mx *sync.RWMutex
	mp map[int64]user.User
}

func (d *BasicCustomer) Find(ctx context.Context, userID int64) (user.User, bool) {
	d.mx.RLock()
	defer d.mx.RUnlock()
	if _, ok := d.mp[userID]; !ok {
		return user.User{}, false
	}
	return d.mp[userID], true
}

func (d *BasicCustomer) ChangeInfo(ctx context.Context, userID int64, nickname, email string) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	cur := d.mp[userID]
	cur.Nickname = nickname
	cur.Email = email
	d.mp[userID] = cur
	return nil
}

func (d *BasicCustomer) CreateByID(ctx context.Context, nickname string, email string, userID int64) (user.User, error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	d.mp[userID] = user.User{ID: userID, Nickname: nickname, Email: email}
	return d.mp[userID], nil
}

func (d *BasicCustomer) DeleteByID(ctx context.Context, userID int64) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	delete(d.mp, userID)
	return nil
}
