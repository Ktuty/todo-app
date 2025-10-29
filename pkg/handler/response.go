package handler

import (
	"encoding/json"
	"errors"
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

var ErrInvalidAction = errors.New("invalid action")

// GetCurrentTime возвращает текущее время в UTC
func GetCurrentTime() time.Time {
	return time.Now().UTC()
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

type RabbitMQRequest struct {
	ID      string          `json:"id"`
	Version string          `json:"version"`
	Action  string          `json:"action"`
	Data    json.RawMessage `json:"data"`
	Auth    string          `json:"auth"`
}

// RabbitMQResponse структура исходящего сообщения в RabbitMQ
type RabbitMQResponse struct {
	CorrelationID string      `json:"correlation_id"`
	Status        string      `json:"status"`
	Data          interface{} `json:"data"`
	Error         *string     `json:"error"`
}

// CreateUserRequest структура для создания пользователя через RabbitMQ
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateListRequest структура для создания списка через RabbitMQ
type CreateListRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateItemRequest структура для создания item через RabbitMQ
type CreateItemRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ListID      int    `json:"list_id"`
}
