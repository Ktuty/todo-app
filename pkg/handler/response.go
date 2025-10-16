package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type errorResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

type rateLimitResponse struct {
	Error      string `json:"error"`
	RetryAfter int    `json:"retry_after,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Remaining  int    `json:"remaining,omitempty"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}

func newRateLimitResponse(c *gin.Context, retryAfter int, limit int, remaining int) {
	c.AbortWithStatusJSON(429, rateLimitResponse{
		Error:      "Rate limit exceeded",
		RetryAfter: retryAfter,
		Limit:      limit,
		Remaining:  remaining,
	})
}

// Дополнительные структуры для ответа
type healthResponse struct {
	Status       string       `json:"status"`
	Timestamp    string       `json:"timestamp"`
	Version      string       `json:"version"`
	Dependencies dependencies `json:"dependencies"`
	System       systemInfo   `json:"system"`
}

type dependencies struct {
	Database string `json:"database"`
	Redis    string `json:"redis,omitempty"`
}

type systemInfo struct {
	Goroutines int    `json:"goroutines"`
	Memory     uint64 `json:"memory_mb"`
	Uptime     string `json:"uptime"`
}

var startTime = time.Now()
