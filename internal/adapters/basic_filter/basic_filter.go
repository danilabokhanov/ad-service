package basicfilter

import (
	"adservice/internal/adpattern"
	"adservice/internal/app"
	"context"
	"sync"
	"time"
)

type BasicFilter struct {
	mx      *sync.RWMutex
	pattern adpattern.AdPattern
}

func (d *BasicFilter) BasicConfig(ctx context.Context) (app.Filter, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.pattern.IsLTimeSet = false
	d.pattern.IsRTimeSet = false
	d.pattern.PublishedOnly = true
	return d, nil
}

func (d *BasicFilter) SetStatus(ctx context.Context, publishedOnly bool) (app.Filter, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.pattern.PublishedOnly = publishedOnly
	return d, nil
}

func (d *BasicFilter) SetAuthor(ctx context.Context, userID int64) (app.Filter, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.pattern.AuthorID = userID
	return d, nil
}

func (d *BasicFilter) SetLTime(ctx context.Context, l time.Time) (app.Filter, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.pattern.IsLTimeSet = true
	d.pattern.LDate = l
	return d, nil
}

func (d *BasicFilter) SetRTime(ctx context.Context, r time.Time) (app.Filter, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.pattern.IsRTimeSet = true
	d.pattern.RDate = r
	return d, nil
}

func (d *BasicFilter) GetPattern(ctx context.Context) (adpattern.AdPattern, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	return d.pattern, nil
}
