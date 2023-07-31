package httpgin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"adservice/internal/app"
)

func NewHTTPServer(port string, a app.App) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	handler := gin.New()
	handler.Use(gin.Recovery())
	handler.Use(CustomLogger)
	v1 := handler.Group("/api/v1")
	AppRouter(v1, a)
	s := &http.Server{Addr: port, Handler: handler}

	return s
}
