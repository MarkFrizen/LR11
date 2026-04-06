package main

import (
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const logFile = "/data/log.txt"

func writeTime(c *gin.Context) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "flock: " + err.Error()})
		return
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	now := time.Now().Format(time.RFC3339)
	if _, err := fmt.Fprintf(f, "[%s] %s\n", c.Request.Host, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "time written", "time": now})
}

func readLog(c *gin.Context) {
	f, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"content": ""})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "flock: " + err.Error()})
		return
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, readErr := f.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{"content": string(data)})
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
