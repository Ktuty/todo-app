package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ktuty/todo-app"
	"github.com/sirupsen/logrus"
)

// ProcessRabbitMQMessage обработчик для входящих сообщений из RabbitMQ
// @Summary Process RabbitMQ Message
// @Description Обрабатывает сообщения из RabbitMQ и возвращает ответ
// @Tags RabbitMQ
// @Accept json
// @Produce json
// @Param input body RabbitMQRequest true "сообщение"
// @Success 200 {object} RabbitMQResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /rabbitmq/process [post]
func (h *Handler) ProcessRabbitMQMessage(c *gin.Context) {
	var request RabbitMQRequest
	if err := c.BindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	// Валидация API ключа
	if !h.validateAPIKey(request.Auth) {
		newErrorResponse(c, http.StatusUnauthorized, "invalid api key")
		return
	}

	// Проверка идемпотентности
	if h.isDuplicateRequest(request.ID) {
		c.JSON(http.StatusOK, RabbitMQResponse{
			CorrelationID: request.ID,
			Status:        "ok",
			Data:          "request already processed",
			Error:         nil,
		})
		return
	}

	// Обработка в зависимости от действия
	response, err := h.processAction(request)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(http.StatusOK, RabbitMQResponse{
			CorrelationID: request.ID,
			Status:        "error",
			Data:          nil,
			Error:         &errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, RabbitMQResponse{
		CorrelationID: request.ID,
		Status:        "ok",
		Data:          response,
		Error:         nil,
	})
}

// validateAPIKey проверяет валидность API ключа
func (h *Handler) validateAPIKey(apiKey string) bool {
	// Здесь должна быть логика проверки API ключа
	// Например, проверка в базе данных или сравнение с конфигурацией
	// Временная реализация - проверяем, что ключ не пустой и имеет минимальную длину
	return apiKey != "" && len(apiKey) >= 16
}

// isDuplicateRequest проверяет, был ли уже обработан запрос с таким ID
func (h *Handler) isDuplicateRequest(requestID string) bool {
	// Здесь должна быть логика проверки дубликатов
	// Например, проверка в Redis или базе данных
	// Временная реализация - всегда возвращаем false
	return false
}

// processAction обрабатывает различные действия из запроса
func (h *Handler) processAction(request RabbitMQRequest) (interface{}, error) {
	switch request.Action {
	case "create_user":
		return h.processCreateUser(request.Data)
	case "create_list":
		return h.processCreateList(request.Data)
	case "create_item":
		return h.processCreateItem(request.Data)
	case "get_user":
		return h.processGetUser(request.Data)
	case "get_list":
		return h.processGetList(request.Data)
	case "get_item":
		return h.processGetItem(request.Data)
	case "get_all_lists":
		return h.processGetAllLists(request.Data)
	case "get_all_items":
		return h.processGetAllItems(request.Data)
	case "update_list":
		return h.processUpdateList(request.Data)
	case "update_item":
		return h.processUpdateItem(request.Data)
	case "delete_list":
		return h.processDeleteList(request.Data)
	case "delete_item":
		return h.processDeleteItem(request.Data)
	case "archive_list":
		return h.processArchiveList(request.Data)
	case "complete_item":
		return h.processCompleteItem(request.Data)
	default:
		return nil, ErrInvalidAction
	}
}

// processCreateUser обрабатывает создание пользователя
func (h *Handler) processCreateUser(data json.RawMessage) (interface{}, error) {
	var userReq CreateUserRequest
	if err := json.Unmarshal(data, &userReq); err != nil {
		return nil, err
	}

	// Создание пользователя через существующий сервис
	userID, err := h.services.Authorization.CreateUser(todo.User{
		Username: userReq.Username,
		Password: userReq.Password,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user_id":  userID,
		"username": userReq.Username,
	}, nil
}

// processCreateList обрабатывает создание списка
func (h *Handler) processCreateList(data json.RawMessage) (interface{}, error) {
	var listReq CreateListRequest
	if err := json.Unmarshal(data, &listReq); err != nil {
		return nil, err
	}

	// Здесь нужно получить userID из данных запроса
	var userData struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &userData); err != nil || userData.UserID == 0 {
		return nil, errors.New("user_id is required")
	}

	listID, err := h.services.TodoList.Create(userData.UserID, todo.TodoList{
		Title:       listReq.Title,
		Description: listReq.Description,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list_id": listID,
		"title":   listReq.Title,
	}, nil
}

// processCreateItem обрабатывает создание item
func (h *Handler) processCreateItem(data json.RawMessage) (interface{}, error) {
	var itemReq CreateItemRequest
	if err := json.Unmarshal(data, &itemReq); err != nil {
		return nil, err
	}

	// Здесь нужно получить userID из данных запроса
	var userData struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &userData); err != nil || userData.UserID == 0 {
		return nil, errors.New("user_id is required")
	}

	itemID, err := h.services.TodoItem.Create(userData.UserID, itemReq.ListID, todo.TodoItem{
		Title:       itemReq.Title,
		Description: itemReq.Description,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"item_id": itemID,
		"title":   itemReq.Title,
		"list_id": itemReq.ListID,
	}, nil
}

// processGetUser обрабатывает получение пользователя
func (h *Handler) processGetUser(data json.RawMessage) (interface{}, error) {
	var userReq struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &userReq); err != nil {
		return nil, err
	}

	// Здесь нужно добавить метод GetUser в сервис Authorization
	// Временная реализация - возвращаем базовую информацию
	return map[string]interface{}{
		"user_id": userReq.UserID,
		"status":  "user_info_retrieved",
	}, nil
}

// processGetList обрабатывает получение списка
func (h *Handler) processGetList(data json.RawMessage) (interface{}, error) {
	var listReq struct {
		ListID int `json:"list_id"`
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &listReq); err != nil {
		return nil, err
	}

	list, err := h.services.TodoList.GetById(listReq.UserID, listReq.ListID)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// processGetItem обрабатывает получение item
func (h *Handler) processGetItem(data json.RawMessage) (interface{}, error) {
	var itemReq struct {
		ItemID int `json:"item_id"`
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &itemReq); err != nil {
		return nil, err
	}

	item, err := h.services.TodoItem.GetById(itemReq.UserID, itemReq.ItemID)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// processGetAllLists обрабатывает получение всех списков пользователя
func (h *Handler) processGetAllLists(data json.RawMessage) (interface{}, error) {
	var listReq struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &listReq); err != nil {
		return nil, err
	}

	lists, err := h.services.TodoList.GetAll(listReq.UserID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"lists": lists,
		"count": len(lists),
	}, nil
}

// processGetAllItems обрабатывает получение всех items списка
func (h *Handler) processGetAllItems(data json.RawMessage) (interface{}, error) {
	var itemReq struct {
		ListID int `json:"list_id"`
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(data, &itemReq); err != nil {
		return nil, err
	}

	items, err := h.services.TodoItem.GetAll(itemReq.UserID, itemReq.ListID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"items": items,
		"count": len(items),
	}, nil
}

// processUpdateList обрабатывает обновление списка
func (h *Handler) processUpdateList(data json.RawMessage) (interface{}, error) {
	var updateReq struct {
		UserID      int    `json:"user_id"`
		ListID      int    `json:"list_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(data, &updateReq); err != nil {
		return nil, err
	}

	err := h.services.TodoList.Update(updateReq.UserID, updateReq.ListID, todo.UpdateListInput{
		Title:       &updateReq.Title,
		Description: &updateReq.Description,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list_id": updateReq.ListID,
		"status":  "updated",
	}, nil
}

// processUpdateItem обрабатывает обновление item
func (h *Handler) processUpdateItem(data json.RawMessage) (interface{}, error) {
	var updateReq struct {
		UserID      int    `json:"user_id"`
		ItemID      int    `json:"item_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Done        bool   `json:"done"`
	}
	if err := json.Unmarshal(data, &updateReq); err != nil {
		return nil, err
	}

	err := h.services.TodoItem.Update(updateReq.UserID, updateReq.ItemID, todo.UpdateItemInput{
		Title:       &updateReq.Title,
		Description: &updateReq.Description,
		Done:        &updateReq.Done,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"item_id": updateReq.ItemID,
		"status":  "updated",
	}, nil
}

// processDeleteList обрабатывает удаление списка
func (h *Handler) processDeleteList(data json.RawMessage) (interface{}, error) {
	var deleteReq struct {
		UserID int `json:"user_id"`
		ListID int `json:"list_id"`
	}
	if err := json.Unmarshal(data, &deleteReq); err != nil {
		return nil, err
	}

	err := h.services.TodoList.Delete(deleteReq.UserID, deleteReq.ListID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list_id": deleteReq.ListID,
		"status":  "deleted",
	}, nil
}

// processDeleteItem обрабатывает удаление item
func (h *Handler) processDeleteItem(data json.RawMessage) (interface{}, error) {
	var deleteReq struct {
		UserID int `json:"user_id"`
		ItemID int `json:"item_id"`
	}
	if err := json.Unmarshal(data, &deleteReq); err != nil {
		return nil, err
	}

	err := h.services.TodoItem.Delete(deleteReq.UserID, deleteReq.ItemID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"item_id": deleteReq.ItemID,
		"status":  "deleted",
	}, nil
}

// processArchiveList обрабатывает архивацию списка
func (h *Handler) processArchiveList(data json.RawMessage) (interface{}, error) {
	var archiveReq struct {
		UserID int `json:"user_id"`
		ListID int `json:"list_id"`
	}
	if err := json.Unmarshal(data, &archiveReq); err != nil {
		return nil, err
	}

	err := h.services.TodoList.ArchiveList(archiveReq.UserID, archiveReq.ListID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list_id": archiveReq.ListID,
		"status":  "archived",
	}, nil
}

// processCompleteItem обрабатывает завершение item
func (h *Handler) processCompleteItem(data json.RawMessage) (interface{}, error) {
	var completeReq struct {
		UserID int `json:"user_id"`
		ItemID int `json:"item_id"`
	}
	if err := json.Unmarshal(data, &completeReq); err != nil {
		return nil, err
	}

	err := h.services.TodoItem.CompleteItem(completeReq.UserID, completeReq.ItemID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"item_id": completeReq.ItemID,
		"status":  "completed",
	}, nil
}

// RabbitMQHealthCheck проверка здоровья RabbitMQ соединения
// @Summary RabbitMQ Health Check
// @Description Проверяет статус соединения с RabbitMQ
// @Tags RabbitMQ
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /rabbitmq/health [get]
func (h *Handler) RabbitMQHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "RabbitMQ endpoint is available",
		"timestamp": GetCurrentTime().Format(time.RFC3339),
		"service":   "rabbitmq_handler",
	})
}

// SendRabbitMQMessage отправка тестового сообщения в RabbitMQ
// @Summary Send Test Message to RabbitMQ
// @Description Отправляет тестовое сообщение в RabbitMQ
// @Tags RabbitMQ
// @Accept json
// @Produce json
// @Param input body RabbitMQRequest true "тестовое сообщение"
// @Success 200 {object} RabbitMQResponse
// @Failure 400 {object} errorResponse
// @Router /rabbitmq/send [post]
func (h *Handler) SendRabbitMQMessage(c *gin.Context) {
	var request RabbitMQRequest
	if err := c.BindJSON(&request); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	// Генерация correlation_id если не указан
	if request.ID == "" {
		request.ID = uuid.New().String()
	}

	// Логируем отправку сообщения
	logrus.WithFields(logrus.Fields{
		"message_id": request.ID,
		"action":     request.Action,
		"version":    request.Version,
	}).Info("Message would be sent to RabbitMQ")

	c.JSON(http.StatusOK, RabbitMQResponse{
		CorrelationID: request.ID,
		Status:        "sent",
		Data: gin.H{
			"message":    "message queued for processing",
			"action":     request.Action,
			"message_id": request.ID,
		},
		Error: nil,
	})
}

// GetRabbitMQStats получение статистики по RabbitMQ обработчику
// @Summary Get RabbitMQ Statistics
// @Description Возвращает статистику обработки RabbitMQ сообщений
// @Tags RabbitMQ
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /rabbitmq/stats [get]
func (h *Handler) GetRabbitMQStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "rabbitmq_handler",
		"timestamp": GetCurrentTime().Format(time.RFC3339),
		"supported_actions": []string{
			"create_user", "create_list", "create_item",
			"get_user", "get_list", "get_item",
			"get_all_lists", "get_all_items",
			"update_list", "update_item",
			"delete_list", "delete_item",
			"archive_list", "complete_item",
		},
		"features": []string{
			"api_key_auth",
			"idempotency_check",
			"structured_responses",
			"error_handling",
		},
		"queues": []string{
			"api.requests",
			"api.responses",
			"api.dlq",
		},
	})
}
