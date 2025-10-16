package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ktuty/todo-app/pkg/handler/middleware"
	"github.com/ktuty/todo-app/pkg/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/ktuty/todo-app/docs"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	// Rate limiter: 100 запросов в минуту для всех эндпоинтов
	rateLimiter := middleware.NewRateLimiter(100, 100)
	router.Use(rateLimiter.RateLimit())

	// Swagger документация
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes - аутентификация
	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	// Версия 1 API - базовая функциональность
	v1 := router.Group("/api/v1", h.userIdentity)
	{
		h.initListRoutes(v1)
		h.initItemRoutes(v1)
	}

	// Версия 2 API - расширенная функциональность с дополнительным rate limiting
	v2 := router.Group("/api/v2", h.userIdentity, h.rateLimit)
	{
		h.initListRoutesV2(v2)
		h.initItemRoutesV2(v2)
	}

	router.GET("/health", h.healthCheck)

	return router
}

func (h *Handler) initListRoutes(api *gin.RouterGroup) {
	lists := api.Group("/lists")
	{
		lists.POST("/", h.createList)
		lists.GET("/", h.getAllLists)
		lists.GET("/:id", h.getListById)
		lists.PUT("/:id", h.updateList)
		lists.DELETE("/:id", h.deleteList)

		items := lists.Group(":id/items")
		{
			items.POST("/", h.createItem)
			items.GET("/", h.getAllItems)
		}
	}
}

func (h *Handler) initItemRoutes(api *gin.RouterGroup) {
	items := api.Group("items")
	{
		items.GET("/:id", h.getItemById)
		items.PUT("/:id", h.updateItem)
		items.DELETE("/:id", h.deleteItem)
	}
}

// V2 routes с новыми возможностями
func (h *Handler) initListRoutesV2(api *gin.RouterGroup) {
	lists := api.Group("/lists")
	{
		lists.POST("/", h.createListV2)            // с поддержкой идемпотентности
		lists.GET("/", h.getAllListsV2)            // с пагинацией и фильтрацией
		lists.GET("/:id", h.getListByIdV2)         // с расширенной информацией
		lists.PUT("/:id", h.updateListV2)          // с частичным обновлением
		lists.DELETE("/:id", h.deleteListV2)       // с мягким удалением
		lists.PATCH("/:id/archive", h.archiveList) // новая возможность - архивация
	}
}

func (h *Handler) initItemRoutesV2(api *gin.RouterGroup) {
	items := api.Group("items")
	{
		items.POST("/", h.createItemV2)              // с поддержкой идемпотентности
		items.GET("/", h.getAllItemsV2)              // с пагинацией
		items.GET("/:id", h.getItemByIdV2)           // с расширенной информацией
		items.PUT("/:id", h.updateItemV2)            // с частичным обновлением
		items.DELETE("/:id", h.deleteItemV2)         // с мягким удалением
		items.PATCH("/:id/complete", h.completeItem) // новая возможность - отметка выполнения
	}
}
