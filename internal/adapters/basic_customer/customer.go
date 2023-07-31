package basiccustomer

import (
	"adservice/internal/app"
	"adservice/internal/user"
	"sync"
)

func New() app.Users {
	return &BasicCustomer{mx: &sync.RWMutex{}, mp: map[int64]user.User{}}
}
