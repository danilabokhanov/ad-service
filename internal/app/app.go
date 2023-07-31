package app

import (
	"adservice/internal/adpattern"
	"adservice/internal/ads"
	"adservice/internal/user"
	"fmt"
	"github.com/danilabokhanov/strintvalidator"
	"golang.org/x/net/context"
	"time"
)

type App interface {
	FindAd(ctx context.Context, adID int64) (ads.Ad, error)
	CreateAd(ctx context.Context, title string, text string, userID int64) (ads.Ad, error)
	DeleteAd(ctx context.Context, adID, userID int64) (ads.Ad, error)
	ChangeAdStatus(ctx context.Context, adID int64, userID int64, published bool) (ads.Ad, error)
	UpdateAd(ctx context.Context, adID int64, userID int64, title string, text string) (ads.Ad, error)
	GetAdsByTitle(ctx context.Context, title string) ([]ads.Ad, error)
	GetAllAdsByTemplate(ctx context.Context, adp adpattern.AdPattern) ([]ads.Ad, error)
	GetNewFilter(ctx context.Context) (Filter, error)
	FindUser(ctx context.Context, userID int64) (user.User, bool, error)
	CreateUserByID(ctx context.Context, nickname, email string, userID int64) (user.User, error)
	DeleteUserByID(ctx context.Context, userID int64) (user.User, error)
	ChangeUserInfo(ctx context.Context, userID int64, nickname, email string) (user.User, error)
}

type Repository interface {
	Find(ctx context.Context, adID int64) (ads.Ad, bool)
	GetByTitle(ctx context.Context, title string) ([]ads.Ad, error)
	Add(ctx context.Context, title string, text string, userID int64) (int64, error)
	Delete(ctx context.Context, adID int64) error
	DeleteByAuthor(ctx context.Context, userID int64) error
	SetTitle(ctx context.Context, adID int64, title string) error
	SetText(ctx context.Context, adID int64, text string) error
	SetStatus(ctx context.Context, adID int64, status bool) error
	GetAllByTemplate(ctx context.Context, adp adpattern.AdPattern) ([]ads.Ad, error)
}

type Users interface {
	Find(ctx context.Context, userID int64) (user.User, bool)
	CreateByID(ctx context.Context, nickname, email string, userID int64) (user.User, error)
	DeleteByID(ctx context.Context, userID int64) error
	ChangeInfo(ctx context.Context, userID int64, nickname, email string) error
}

type Filter interface {
	BasicConfig(ctx context.Context) (Filter, error)
	SetStatus(ctx context.Context, publishedOnly bool) (Filter, error)
	SetAuthor(ctx context.Context, userID int64) (Filter, error)
	SetLTime(ctx context.Context, l time.Time) (Filter, error)
	SetRTime(ctx context.Context, r time.Time) (Filter, error)
	GetPattern(ctx context.Context) (adpattern.AdPattern, error)
}

type SimpleApp struct {
	repository Repository
	users      Users
	filter     Filter
}

func NewApp(repo Repository, u Users, f Filter) App {
	return &SimpleApp{repository: repo, users: u, filter: f}
}

var ErrWrongFormat = fmt.Errorf("wrong format")
var ErrNoAccess = fmt.Errorf("permission denied")
var ErrApp = fmt.Errorf("unknown application error")
var ErrUserDoesNotExist = fmt.Errorf("user doesn't exist")

func (d SimpleApp) CreateAd(ctx context.Context, title string, text string, userID int64) (ads.Ad, error) {
	if e := strintvalidator.Validate(ads.Ad{Title: title, Text: text}); e != nil {
		return ads.Ad{}, ErrWrongFormat
	}
	_, isFound := d.users.Find(ctx, userID)
	if !isFound {
		return ads.Ad{}, ErrUserDoesNotExist
	}
	adID, err := d.repository.Add(ctx, title, text, userID)
	if err != nil {
		return ads.Ad{}, ErrApp
	}
	ad, _ := d.repository.Find(ctx, adID)
	return ad, nil
}

func (d SimpleApp) DeleteAd(ctx context.Context, adID, userID int64) (ads.Ad, error) {
	ad, isFound := d.repository.Find(ctx, adID)
	if !isFound {
		return ads.Ad{}, ErrWrongFormat
	}
	if ad.AuthorID != userID {
		return ads.Ad{}, ErrNoAccess
	}
	err := d.repository.Delete(ctx, adID)
	if err != nil {
		return ads.Ad{}, ErrApp
	}
	return ad, nil
}

func (d SimpleApp) ChangeAdStatus(ctx context.Context, adID int64, userID int64, published bool) (ads.Ad, error) {
	_, isFound := d.users.Find(ctx, userID)
	if !isFound {
		return ads.Ad{}, ErrUserDoesNotExist
	}
	ad, isFound := d.repository.Find(ctx, adID)
	if !isFound {
		return ads.Ad{}, ErrWrongFormat
	}
	if ad.AuthorID != userID {
		return ads.Ad{}, ErrNoAccess
	}
	err := d.repository.SetStatus(ctx, adID, published)
	if err != nil {
		return ads.Ad{}, ErrApp
	}
	ad.Published = published
	return ad, nil
}

func (d SimpleApp) UpdateAd(ctx context.Context, adID int64, userID int64, title string, text string) (ads.Ad, error) {
	if e := strintvalidator.Validate(ads.Ad{Title: title, Text: text}); e != nil {
		return ads.Ad{}, ErrWrongFormat
	}
	_, isFound := d.users.Find(ctx, userID)
	if !isFound {
		return ads.Ad{}, ErrUserDoesNotExist
	}
	ad, isFound := d.repository.Find(ctx, adID)
	if !isFound {
		return ads.Ad{}, ErrWrongFormat
	}
	if ad.AuthorID != userID {
		return ads.Ad{}, ErrNoAccess
	}
	err := d.repository.SetText(ctx, adID, text)
	if err != nil {
		return ads.Ad{}, ErrApp
	}
	err = d.repository.SetTitle(ctx, adID, title)
	if err != nil {
		return ads.Ad{}, ErrApp
	}
	ad.Title = title
	ad.Text = text
	return ad, nil
}

func (d SimpleApp) ChangeUserInfo(ctx context.Context, userID int64, nickname, email string) (user.User, error) {
	u, isFound := d.users.Find(ctx, userID)
	if !isFound {
		return user.User{}, ErrUserDoesNotExist
	}
	err := d.users.ChangeInfo(ctx, userID, nickname, email)
	if err != nil {
		return user.User{}, ErrApp
	}
	u.Nickname = nickname
	u.Email = email
	return u, nil
}

func (d SimpleApp) GetAllAdsByTemplate(ctx context.Context, adp adpattern.AdPattern) ([]ads.Ad, error) {
	res, err := d.repository.GetAllByTemplate(ctx, adp)
	if err != nil {
		return []ads.Ad{}, ErrApp
	}
	return res, nil
}

func (d SimpleApp) GetNewFilter(ctx context.Context) (Filter, error) {
	f, err := d.filter.BasicConfig(ctx)
	if err != nil {
		return f, ErrApp
	}
	return f, nil
}

func CheckAd(ad ads.Ad, pattern adpattern.AdPattern) bool {
	if pattern.PublishedOnly && !ad.Published {
		return false
	}
	if pattern.AuthorID != 0 && pattern.AuthorID != ad.AuthorID {
		return false
	}

	if pattern.IsRTimeSet && pattern.RDate.Before(ad.CreationDate) ||
		pattern.IsLTimeSet && pattern.LDate.After(ad.CreationDate) {
		return false
	}
	return true
}

func (d SimpleApp) GetAdsByTitle(ctx context.Context, title string) ([]ads.Ad, error) {
	res, err := d.repository.GetByTitle(ctx, title)
	if err != nil {
		return []ads.Ad{}, ErrApp
	}
	return res, nil
}

func (d SimpleApp) FindAd(ctx context.Context, adID int64) (ads.Ad, error) {
	ad, isFound := d.repository.Find(ctx, adID)
	if !isFound {
		return ads.Ad{}, ErrWrongFormat
	}
	return ad, nil
}

func (d SimpleApp) CreateUserByID(ctx context.Context, nickname, email string, userID int64) (user.User, error) {
	_, isFound := d.users.Find(ctx, userID)
	if isFound {
		return user.User{}, ErrNoAccess
	}
	u, err := d.users.CreateByID(ctx, nickname, email, userID)
	if err != nil {
		return user.User{}, ErrApp
	}
	return u, nil
}

func (d SimpleApp) DeleteUserByID(ctx context.Context, userID int64) (user.User, error) {
	u, isFound := d.users.Find(ctx, userID)
	if !isFound {
		return user.User{}, ErrUserDoesNotExist
	}
	err := d.users.DeleteByID(ctx, userID)
	if err != nil {
		return user.User{}, ErrApp
	}
	err = d.repository.DeleteByAuthor(ctx, userID)
	if err != nil {
		return user.User{}, ErrApp
	}
	return u, nil
}

func (d SimpleApp) FindUser(ctx context.Context, userID int64) (user.User, bool, error) {
	u, isFound := d.users.Find(ctx, userID)
	return u, isFound, nil
}
