package maprepo

import (
	"adservice/internal/ads"
	"adservice/internal/app"
	"sync"
)

func New() app.Repository {
	return &MapRepo{mx: &sync.RWMutex{}, mp: map[int64]ads.Ad{}}
}
