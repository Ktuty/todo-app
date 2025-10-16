package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"time"
)

// HealthCheck возвращает статус сервиса и его зависимостей
// @Summary Health check
// @Description Check service health status
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} healthResponse
// @Router /health [get]
func (h *Handler) healthCheck(c *gin.Context) {
	// Проверяем статус базы данных
	dbStatus := "connected"
	if err := h.checkDatabase(); err != nil {
		dbStatus = "disconnected"
	}

	// Получаем информацию о системе
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, healthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "2.0",
		Dependencies: dependencies{
			Database: dbStatus,
		},
		System: systemInfo{
			Goroutines: runtime.NumGoroutine(),
			Memory:     m.Alloc / 1024 / 1024, // MB
			Uptime:     time.Since(startTime).String(),
		},
	})
}

// checkDatabase проверяет подключение к базе данных
func (h *Handler) checkDatabase() error {
	// возвращаем nil, проверка не реализована
	return nil
}
