package handler

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// StartRabbitMQConsumer запускает consumer для обработки сообщений из RabbitMQ
func (h *Handler) StartRabbitMQConsumer() error {
	logrus.Info("Starting RabbitMQ consumer...")

	err := h.rabbitMQ.ConsumeRequests(func(delivery amqp.Delivery) error {
		return h.processRabbitMQMessage(delivery)
	})

	if err != nil {
		return fmt.Errorf("failed to start RabbitMQ consumer: %w", err)
	}

	logrus.Info("RabbitMQ consumer started successfully")
	return nil
}

// processRabbitMQMessage обрабатывает одно сообщение из RabbitMQ
func (h *Handler) processRabbitMQMessage(delivery amqp.Delivery) error {
	correlationID := delivery.CorrelationId
	if correlationID == "" {
		correlationID = delivery.MessageId
	}

	replyTo := delivery.ReplyTo

	logrus.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"message_id":     delivery.MessageId,
		"reply_to":       replyTo,
		"timestamp":      delivery.Timestamp,
	}).Info("Received message from RabbitMQ")

	var request RabbitMQRequest
	if err := json.Unmarshal(delivery.Body, &request); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal RabbitMQ message")
		h.sendErrorResponse(correlationID, "invalid message format", replyTo)
		return err
	}

	// Если в сообщении нет ID, используем correlation ID
	if request.ID == "" {
		request.ID = correlationID
	}

	// Валидация API ключа
	if !h.validateAPIKey(request.Auth) {
		logrus.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"api_key":        request.Auth,
		}).Warn("Invalid API key")
		h.sendErrorResponse(correlationID, "invalid api key", replyTo)
		return fmt.Errorf("invalid api key")
	}

	// Проверка идемпотентности
	if h.isDuplicateRequest(request.ID) {
		logrus.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"request_id":     request.ID,
		}).Info("Duplicate request detected")

		h.rabbitMQ.PublishResponse(correlationID, map[string]interface{}{
			"message": "request already processed",
			"status":  "duplicate",
		}, nil, replyTo)
		return nil
	}

	// Обработка действия
	response, err := h.processAction(request)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"action":         request.Action,
			"error":          err.Error(),
		}).Error("Failed to process action")

		h.sendErrorResponse(correlationID, err.Error(), replyTo)
		return err
	}

	// Отправляем успешный ответ
	if err := h.rabbitMQ.PublishResponse(correlationID, response, nil, replyTo); err != nil {
		logrus.WithError(err).Error("Failed to publish response")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"action":         request.Action,
		"reply_to":       replyTo,
		"status":         "success",
	}).Info("Message processed successfully")

	return nil
}

// sendErrorResponse отправляет ответ с ошибкой
func (h *Handler) sendErrorResponse(correlationID string, errorMsg string, replyTo string) {
	if err := h.rabbitMQ.PublishResponse(correlationID, nil, fmt.Errorf(errorMsg), replyTo); err != nil {
		logrus.WithError(err).Error("Failed to publish error response")
	}
}
