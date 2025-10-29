package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	// Подключаемся к RabbitMQ
	conn, err := amqp.Dial("amqp://admin:password@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open channel:", err)
	}
	defer ch.Close()

	// Создаем временную очередь для ответов
	replyQueue, err := ch.QueueDeclare(
		"",    // name - пустое для автоматической генерации
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatal("Failed to declare reply queue:", err)
	}

	// Потребляем ответы из нашей временной очереди
	msgs, err := ch.Consume(
		replyQueue.Name, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatal("Failed to consume responses:", err)
	}

	// Генерируем уникальное имя пользователя
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("testuser_%d", timestamp)

	// Создаем тестовое сообщение
	request := map[string]interface{}{
		"id":      fmt.Sprintf("test-%d", timestamp),
		"version": "v1",
		"action":  "create_user",
		"data": map[string]interface{}{
			"username": username,
			"password": "testpass123",
		},
		"auth": "todo-app-api-key-12345",
	}

	body, _ := json.Marshal(request)

	correlationID := fmt.Sprintf("test-corr-%d", timestamp)

	fmt.Printf("Sending message with correlation_id: %s\n", correlationID)
	fmt.Printf("Waiting for response in queue: %s\n", replyQueue.Name)
	fmt.Printf("Creating user: %s\n", username)

	// Публикуем запрос
	err = ch.Publish(
		"",             // exchange
		"api.requests", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyQueue.Name, // Указываем куда отправлять ответ
			Body:          body,
			Timestamp:     time.Now(),
		},
	)
	if err != nil {
		log.Fatal("Failed to publish message:", err)
	}

	fmt.Println("Message sent, waiting for response...")

	// Ждем ответ с таймаутом
	select {
	case msg := <-msgs:
		fmt.Printf("✅ Received response:\n%s\n", string(msg.Body))

		// Парсим ответ для красивого вывода
		var response map[string]interface{}
		if err := json.Unmarshal(msg.Body, &response); err == nil {
			fmt.Printf("\n📊 Response details:\n")
			fmt.Printf("   Status: %s\n", response["status"])
			fmt.Printf("   Correlation ID: %s\n", response["correlation_id"])
			fmt.Printf("   Timestamp: %s\n", response["timestamp"])

			if data, ok := response["data"].(map[string]interface{}); ok {
				fmt.Printf("   Data:\n")
				for key, value := range data {
					fmt.Printf("      %s: %v\n", key, value)
				}
			}
			if errorMsg, ok := response["error"].(string); ok && errorMsg != "" {
				fmt.Printf("   ❌ Error: %s\n", errorMsg)
			}
		}

	case <-time.After(10 * time.Second):
		fmt.Println("❌ Timeout waiting for response")
	}
}
