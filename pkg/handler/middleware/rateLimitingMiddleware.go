package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.limiter.Allow() {
			retryAfter := time.Now().Add(time.Minute).Unix()

			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limiter.Burst()))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(retryAfter, 10))
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}

		// Добавляем заголовки с информацией о лимитах
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limiter.Burst()))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.limiter.Burst()-1))
		c.Next()
	}
}
