package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ktuty/todo-app"
)

// CreateListV2 создает новый список с поддержкой идемпотентности
// @Summary Create todo list (v2)
// @Description Create todo list with idempotency support
// @Security ApiKeyAuth
// @Tags lists-v2
// @Accept json
// @Produce json
// @Param input body createListV2Request true "List info with idempotency key"
// @Success 201 {object} todo.TodoList
// @Success 200 {object} idempotentResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/lists [post]
func (h *Handler) createListV2(c *gin.Context) {
	var input createListV2Request

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Проверка идемпотентности
	if input.IdempotencyKey != "" {
		idempotencyKey := generateIdempotencyKey(c.GetHeader("Authorization"), input.IdempotencyKey)

		// Проверяем, был ли уже обработан такой запрос
		existingId, err := h.services.CheckIdempotency(userId, idempotencyKey)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "Internal server error")
			return
		}

		if existingId > 0 {
			// Возвращаем ранее созданный ресурс
			list, err := h.services.TodoList.GetById(userId, existingId)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, idempotentResponse{
				Message: "Resource already created",
				Status:  "duplicate",
				Data:    list,
			})
			return
		}
	}

	list := todo.TodoList{
		Title:       input.Title,
		Description: input.Description,
	}

	id, err := h.services.TodoList.Create(userId, list)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Сохраняем ключ идемпотентности если он был передан
	if input.IdempotencyKey != "" {
		idempotencyKey := generateIdempotencyKey(c.GetHeader("Authorization"), input.IdempotencyKey)
		err = h.services.StoreIdempotency(userId, idempotencyKey, id, 24*time.Hour)
		if err != nil {
			// Логируем ошибку, но не прерываем выполнение
			c.JSON(http.StatusCreated, todo.TodoList{
				Id:          id,
				Title:       input.Title,
				Description: input.Description,
			})
			return
		}
	}

	list.Id = id
	c.JSON(http.StatusCreated, list)
}

// GetAllListsV2 получает все списки с пагинацией
// @Summary Get all lists (v2)
// @Description Get all todo lists with pagination and filtering
// @Security ApiKeyAuth
// @Tags lists-v2
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param archived query bool false "Filter by archived status"
// @Success 200 {object} getAllListsV2Response
// @Failure 500 {object} errorResponse
// @Router /api/v2/lists [get]
func (h *Handler) getAllListsV2(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Получаем параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	archived := c.Query("archived")

	// Валидация параметров
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Получаем списки с пагинацией
	lists, total, err := h.services.TodoList.GetAllWithPagination(userId, offset, limit, archived)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Добавляем заголовки с информацией о пагинации
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("X-Page", strconv.Itoa(page))
	c.Header("X-Limit", strconv.Itoa(limit))

	c.JSON(http.StatusOK, getAllListsV2Response{
		Data: lists,
		Meta: paginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: (total + limit - 1) / limit, // Расчет общего количества страниц
		},
	})
}

// GetListByIdV2 получает список по ID с расширенной информацией
// @Summary Get list by ID (v2)
// @Description Get todo list by ID with extended information including item count
// @Security ApiKeyAuth
// @Tags lists-v2
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Success 200 {object} todoListV2Response
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/lists/{id} [get]
func (h *Handler) getListByIdV2(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	list, err := h.services.TodoList.GetById(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Получаем дополнительную информацию для v2
	itemCount, err := h.services.TodoList.GetItemCount(userId, id)
	if err != nil {
		// Если не удалось получить количество items, продолжаем без этой информации
		itemCount = 0
	}

	response := todoListV2Response{
		TodoList:  list,
		ItemCount: itemCount,
		CreatedAt: time.Now(), // В реальной реализации это должно браться из БД
		UpdatedAt: time.Now(), // В реальной реализации это должно браться из БД
	}

	c.JSON(http.StatusOK, response)
}

// UpdateListV2 обновляет список с частичным обновлением
// @Summary Update list (v2)
// @Description Update todo list with partial update support
// @Security ApiKeyAuth
// @Tags lists-v2
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Param input body updateListV2Request true "List update data"
// @Success 200 {object} todo.TodoList
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/lists/{id} [put]
func (h *Handler) updateListV2(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	var input updateListV2Request
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Преобразуем в стандартный input
	updateInput := todo.UpdateListInput{
		Title:       input.Title,
		Description: input.Description,
		Archived:    input.Archived,
	}

	if err := h.services.TodoList.Update(userId, id, updateInput); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращаем обновленный список
	list, err := h.services.TodoList.GetById(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, list)
}

// DeleteListV2 удаляет список (мягкое удаление)
// @Summary Delete list (v2)
// @Description Delete todo list with soft delete
// @Security ApiKeyAuth
// @Tags lists-v2
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/lists/{id} [delete]
func (h *Handler) deleteListV2(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	// Мягкое удаление - устанавливаем флаг archived
	err = h.services.TodoList.ArchiveList(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "list archived successfully",
	})
}

// ArchiveList архивирует список
// @Summary Archive list
// @Description Archive todo list
// @Security ApiKeyAuth
// @Tags lists-v2
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/lists/{id}/archive [patch]
func (h *Handler) archiveList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.services.TodoList.ArchiveList(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "list archived successfully",
	})
}

// CreateItemV2 создает новый item с поддержкой идемпотентности
// @Summary Create todo item (v2)
// @Description Create todo item with idempotency support
// @Security ApiKeyAuth
// @Tags items-v2
// @Accept json
// @Produce json
// @Param input body createItemV2Request true "Item info with idempotency key"
// @Success 201 {object} todo.TodoItem
// @Success 200 {object} idempotentResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/items [post]
func (h *Handler) createItemV2(c *gin.Context) {
	var input createItemV2Request

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Проверка идемпотентности
	if input.IdempotencyKey != "" {
		idempotencyKey := generateIdempotencyKey(c.GetHeader("Authorization"), input.IdempotencyKey)

		existingId, err := h.services.CheckIdempotency(userId, idempotencyKey)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "Internal server error")
			return
		}

		if existingId > 0 {
			item, err := h.services.TodoItem.GetById(userId, existingId)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, idempotentResponse{
				Message: "Resource already created",
				Status:  "duplicate",
				Data:    item,
			})
			return
		}
	}

	item := todo.TodoItem{
		Title:       input.Title,
		Description: input.Description,
	}

	id, err := h.services.TodoItem.Create(userId, input.ListId, item)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Сохраняем ключ идемпотентности если он был передан
	if input.IdempotencyKey != "" {
		idempotencyKey := generateIdempotencyKey(c.GetHeader("Authorization"), input.IdempotencyKey)
		h.services.StoreIdempotency(userId, idempotencyKey, id, 24*time.Hour)
	}

	item.Id = id
	c.JSON(http.StatusCreated, item)
}

// GetAllItemsV2 получает все items с пагинацией
// @Summary Get all items (v2)
// @Description Get all todo items with pagination
// @Security ApiKeyAuth
// @Tags items-v2
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param list_id query int false "Filter by list ID"
// @Param completed query bool false "Filter by completion status"
// @Success 200 {object} getAllItemsV2Response
// @Failure 500 {object} errorResponse
// @Router /api/v2/items [get]
func (h *Handler) getAllItemsV2(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	listId, _ := strconv.Atoi(c.Query("list_id"))
	completed := c.Query("completed")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	items, total, err := h.services.TodoItem.GetAllWithPagination(userId, listId, offset, limit, completed)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("X-Page", strconv.Itoa(page))
	c.Header("X-Limit", strconv.Itoa(limit))

	c.JSON(http.StatusOK, getAllItemsV2Response{
		Data: items,
		Meta: paginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: (total + limit - 1) / limit,
		},
	})
}

// GetItemByIdV2 получает item по ID
// @Summary Get item by ID (v2)
// @Description Get todo item by ID with extended information
// @Security ApiKeyAuth
// @Tags items-v2
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} todo.TodoItem
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/items/{id} [get]
func (h *Handler) getItemByIdV2(c *gin.Context) {
	h.getItemById(c)
}

// UpdateItemV2 обновляет item
// @Summary Update item (v2)
// @Description Update todo item with partial update support
// @Security ApiKeyAuth
// @Tags items-v2
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param input body todo.UpdateItemInput true "Item update data"
// @Success 200 {object} todo.TodoItem
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/items/{id} [put]
func (h *Handler) updateItemV2(c *gin.Context) {
	h.updateItem(c)
}

// DeleteItemV2 удаляет item (мягкое удаление)
// @Summary Delete item (v2)
// @Description Delete todo item with soft delete
// @Security ApiKeyAuth
// @Tags items-v2
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/items/{id} [delete]
func (h *Handler) deleteItemV2(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.services.TodoItem.ArchiveItem(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "item archived successfully",
	})
}

// CompleteItem отмечает item как выполненный
// @Summary Complete item
// @Description Mark todo item as completed
// @Security ApiKeyAuth
// @Tags items-v2
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v2/items/{id}/complete [patch]
func (h *Handler) completeItem(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.services.TodoItem.CompleteItem(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "item completed successfully",
	})
}

// Вспомогательные структуры для v2 API
type createListV2Request struct {
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type createItemV2Request struct {
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	ListId         int    `json:"list_id" binding:"required"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type updateListV2Request struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Archived    *bool   `json:"archived"`
}

type idempotentResponse struct {
	Message string      `json:"message"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
}

type getAllListsV2Response struct {
	Data []todo.TodoList `json:"data"`
	Meta paginationMeta  `json:"meta"`
}

type getAllItemsV2Response struct {
	Data []todo.TodoItem `json:"data"`
	Meta paginationMeta  `json:"meta"`
}

type todoListV2Response struct {
	todo.TodoList
	ItemCount int       `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type paginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

func generateIdempotencyKey(authHeader, userKey string) string {
	hash := sha256.Sum256([]byte(authHeader + userKey))
	return hex.EncodeToString(hash[:])
}
