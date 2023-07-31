package httpgin

import (
	"adservice/internal/app"
	"github.com/gin-gonic/gin"
)

func AppRouter(r *gin.RouterGroup, a app.App) {
	r.POST("/ads", createAd(a))
	r.PUT("/ads/:ad_id/status", changeAdStatus(a))
	r.PUT("/ads/:ad_id", updateAd(a))
	r.DELETE("/ads/:ad_id", deleteAd(a))
	r.GET("/ads", listAds(a))
	r.GET("/ads/by_title", getAdsByTitle(a))
	r.GET("/ads/:ad_id", getAdByID(a))
	r.POST("/users", createUser(a))
	r.PUT("/users/:user_id", changeUserInfo(a))
	r.GET("/users/:user_id", getUserByID(a))
	r.DELETE("/users/:user_id", deleteUserByID(a))
}
