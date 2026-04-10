package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// WriteTime обрабатывает GET /write
func (h *Handler) WriteTime(c *gin.Context) {
	now := h.timeSource.Now().Format(time.RFC3339)
	line := h.formatter.FormatLogLine(c.Request.Host, now)

	if err := h.storage.WriteLine(line); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "time written", "time": now})
}

// ReadLog обрабатывает GET /read
func (h *Handler) ReadLog(c *gin.Context) {
	content, err := h.storage.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

func main() {
	// Инициализация зависимостей (Dependency Injection)
	storage := NewFileLogStorage("/data/log.txt")
	formatter := &DefaultLogFormatter{}
	timeSource := &RealTimeSource{}
	handler := NewHandler(storage, formatter, timeSource)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "go-gin",
			"status":  "running",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	r.GET("/write", handler.WriteTime)
	r.GET("/read", handler.ReadLog)

	r.Run(":8080")
}
