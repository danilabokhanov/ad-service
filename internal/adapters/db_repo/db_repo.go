package dbrepo

import (
	"adservice/internal/adpattern"
	"adservice/internal/ads"
	"adservice/internal/app"
	"context"
	"sort"
	"strings"
	"time"
)

type DataBase interface {
	SelectById(ctx context.Context, adID int64) ([]ads.Ad, error)
	SelectAllIndexes(ctx context.Context) ([]int64, error)
	SelectAll(ctx context.Context) ([]ads.Ad, error)
	Insert(ctx context.Context, ad ads.Ad) error
	UpdateTitle(ctx context.Context, adID int64, title string, UpdateDate time.Time) error
	UpdateText(ctx context.Context, adID int64, text string, UpdateDate time.Time) error
	UpdateStatus(ctx context.Context, adID int64, status bool, UpdateDate time.Time) error
	DeleteById(ctx context.Context, adID int64) error
	DeleteByAuthor(ctx context.Context, userID int64) error
	Close() error
}

type DBRepo struct {
	db DataBase
}

func (d *DBRepo) Find(ctx context.Context, adID int64) (ads.Ad, bool) {
	ad, err := d.db.SelectById(ctx, adID)
	if err != nil {
		return ads.Ad{}, false
	}
	if len(ad) != 1 {
		return ads.Ad{}, false
	}
	return ad[0], true
}

func (d *DBRepo) Add(ctx context.Context, title string, text string, userID int64) (int64, error) {
	indexes, err := d.db.SelectAllIndexes(ctx)
	if err != nil {
		return 0, err
	}
	id := int64(0)
	iter := 0
	for iter < len(indexes) {
		if id != indexes[iter] {
			break
		}
		id++
		iter++
	}
	err = d.db.Insert(ctx, ads.Ad{ID: id, Title: title, Text: text, AuthorID: userID,
		Published: false, CreationDate: time.Now().UTC(), UpdateDate: time.Now().UTC()})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *DBRepo) SetTitle(ctx context.Context, adID int64, title string) error {
	return d.db.UpdateTitle(ctx, adID, title, time.Now().UTC())
}

func (d *DBRepo) SetText(ctx context.Context, adID int64, text string) error {
	return d.db.UpdateText(ctx, adID, text, time.Now().UTC())
}

func (d *DBRepo) SetStatus(ctx context.Context, adID int64, status bool) error {
	return d.db.UpdateStatus(ctx, adID, status, time.Now().UTC())
}

func (d *DBRepo) GetAllByTemplate(ctx context.Context, adp adpattern.AdPattern) ([]ads.Ad, error) {
	adverts, err := d.db.SelectAll(ctx)
	if err != nil {
		return []ads.Ad{}, err
	}
	res := []ads.Ad{}
	for _, ad := range adverts {
		if app.CheckAd(ad, adp) {
			res = append(res, ad)
		}
	}
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].CreationDate.Before(res[j].CreationDate)
	})
	return res, nil
}

func (d *DBRepo) GetByTitle(ctx context.Context, title string) ([]ads.Ad, error) {
	adverts, err := d.db.SelectAll(ctx)
	if err != nil {
		return []ads.Ad{}, err
	}
	res := []ads.Ad{}
	for _, ad := range adverts {
		if strings.HasPrefix(ad.Title, title) {
			res = append(res, ad)
		}
	}
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].CreationDate.Before(res[j].CreationDate)
	})
	return res, nil
}

func (d *DBRepo) Delete(ctx context.Context, adID int64) error {
	return d.db.DeleteById(ctx, adID)
}

func (d *DBRepo) DeleteByAuthor(ctx context.Context, userID int64) error {
	return d.db.DeleteByAuthor(ctx, userID)
}
