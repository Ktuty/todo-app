package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func testAction(action string, data map[string]interface{}) {
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
		"", false, true, true, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare reply queue:", err)
	}

	// Слушаем ответы
	msgs, err := ch.Consume(
		replyQueue.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to consume responses:", err)
	}

	timestamp := time.Now().Unix()
	request := map[string]interface{}{
		"id":      fmt.Sprintf("test-%d", timestamp),
		"version": "v1",
		"action":  action,
		"data":    data,
		"auth":    "todo-app-api-key-12345",
	}

	body, _ := json.Marshal(request)
	correlationID := fmt.Sprintf("test-%d", timestamp)

	fmt.Printf("\n🚀 Testing action: %s\n", action)
	fmt.Printf("   Correlation ID: %s\n", correlationID)

	err = ch.Publish(
		"", "api.requests", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyQueue.Name,
			Body:          body,
			Timestamp:     time.Now(),
		},
	)
	if err != nil {
		log.Fatal("Failed to publish message:", err)
	}

	select {
	case msg := <-msgs:
		var response map[string]interface{}
		json.Unmarshal(msg.Body, &response)

		if response["status"] == "success" {
			fmt.Printf("   ✅ SUCCESS: %s completed\n", action)
			if data, ok := response["data"].(map[string]interface{}); ok {
				for k, v := range data {
					fmt.Printf("      %s: %v\n", k, v)
				}
			}
		} else {
			fmt.Printf("   ❌ ERROR in %s: %v\n", action, response["error"])
		}

	case <-time.After(10 * time.Second):
		fmt.Printf("   ⏰ TIMEOUT: %s\n", action)
	}
}

func main() {
	fmt.Println("🧪 Testing RabbitMQ Actions")
	fmt.Println("============================")

	// Сначала создаем пользователя
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("user_%d", timestamp)

	testAction("create_user", map[string]interface{}{
		"username": username,
		"password": "password123",
	})

	// Тестируем создание списка (нужен user_id)
	testAction("create_list", map[string]interface{}{
		"user_id":     1, // Предполагаем, что user_id = 1 существует
		"title":       "Test List via RabbitMQ",
		"description": "Created through message queue",
	})

	// Тестируем получение списков
	testAction("get_all_lists", map[string]interface{}{
		"user_id": 1,
	})

}
