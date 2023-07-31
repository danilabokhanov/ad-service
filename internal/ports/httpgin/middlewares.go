package httpgin

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (d bodyWriter) Write(data []byte) (int, error) {
	d.body.Write(data)
	return d.ResponseWriter.Write(data)
}

func CustomLogger(c *gin.Context) {
	t := time.Now()
	path := c.Request.URL.Path
	b := bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = b
	c.Next()

	latency := time.Since(t)
	method := c.Request.Method
	status := c.Writer.Status()
	clientIP := c.ClientIP()
	bodySize := c.Writer.Size()

	log.Println("latency", latency, "method", method, "path", path, "status", status,
		"client_ip", clientIP, "body_size", bodySize)
	if status >= 400 {
		log.Println("body:", b.body.String())
	}
}
