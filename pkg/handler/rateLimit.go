package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ktuty/todo-app/pkg/handler/middleware"
)

// RateLimit middleware для v2 API
func (h *Handler) rateLimit(c *gin.Context) {
	// Создаем rate limiter специально для v2 API (более либеральные лимиты)
	v2RateLimiter := middleware.NewRateLimiter(200, 300) // 200 запросов в минуту, burst 300
	v2RateLimiter.RateLimit()(c)
}
