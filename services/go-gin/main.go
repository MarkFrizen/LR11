package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func writeTime(c *gin.Context) {
	now := formatTime(time.Now())
	line := formatLogLine(c.Request.Host, now)

	if err := writeToFile(logFile, line); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "time written", "time": now})
}

func readLog(c *gin.Context) {
	content, err := readFromFile(logFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

func main() {
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

	r.GET("/write", writeTime)
	r.GET("/read", readLog)

	r.Run(":8080")
}
