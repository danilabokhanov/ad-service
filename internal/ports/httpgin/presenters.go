package httpgin

import (
	"adservice/internal/ads"
	"adservice/internal/user"
	"github.com/gin-gonic/gin"
	"time"
)

type createAdRequest struct {
	Title  string `json:"title" binding:"required"`
	Text   string `json:"text" binding:"required"`
	UserID int64  `json:"user_id" binding:"required"`
}

type deleteAdRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

type universalUser struct {
	Nickname string `json:"nickname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	ID       int64  `json:"user_id" binding:"required"`
}

type adResponse struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Text         string    `json:"text"`
	AuthorID     int64     `json:"author_id"`
	Published    bool      `json:"published"`
	CreationDate time.Time `json:"creation_date"`
	UpdateDate   time.Time `json:"update_date"`
}

type changeAdStatusRequest struct {
	Published bool  `json:"published"`
	UserID    int64 `json:"user_id" binding:"required"`
}

type changeUserStatusRequest struct {
	Nickname string `json:"nickname" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type updateAdRequest struct {
	Title  string `json:"title" binding:"required"`
	Text   string `json:"text" binding:"required"`
	UserID int64  `json:"user_id" binding:"required"`
}

func AdSuccessResponse(ad *ads.Ad) *gin.H {
	return &gin.H{
		"data": adResponse{
			ID:           ad.ID,
			Title:        ad.Title,
			Text:         ad.Text,
			AuthorID:     ad.AuthorID,
			Published:    ad.Published,
			CreationDate: ad.CreationDate,
			UpdateDate:   ad.UpdateDate,
		},
		"error": nil,
	}
}

func UserSuccessResponse(u *user.User) *gin.H {
	return &gin.H{
		"data": universalUser{
			ID:       u.ID,
			Nickname: u.Nickname,
			Email:    u.Email,
		},
		"error": nil,
	}
}

func AdSuccessResponseList(ads *[]ads.Ad) *gin.H {
	res := []adResponse{}
	for _, ad := range *ads {
		res = append(res, adResponse{
			ID:           ad.ID,
			Title:        ad.Title,
			Text:         ad.Text,
			AuthorID:     ad.AuthorID,
			Published:    ad.Published,
			CreationDate: ad.CreationDate,
			UpdateDate:   ad.UpdateDate,
		})
	}
	return &gin.H{
		"data":  res,
		"error": nil,
	}
}

func ErrorResponse(err error) *gin.H {
	return &gin.H{
		"data":  nil,
		"error": err.Error(),
	}
}
