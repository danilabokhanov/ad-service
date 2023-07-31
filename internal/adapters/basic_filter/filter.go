package basicfilter

import (
	"adservice/internal/app"
	"sync"
)

func New() app.Filter {
	return &BasicFilter{mx: &sync.RWMutex{}}
}
