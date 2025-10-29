package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type ConfigMQ struct {
	URL string
}

type ServiceMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  ConfigMQ
}

func NewRabbitMQService(config ConfigMQ) (*ServiceMQ, error) {
	service := &ServiceMQ{config: config}
	if err := service.connect(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *ServiceMQ) Close() {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *ServiceMQ) connect() error {
	var err error
	s.conn, err = amqp.Dial(s.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	s.channel, err = s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queues
	if err := s.declareQueues(); err != nil {
		return err
	}

	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// DLQ
func (s *ServiceMQ) declareQueues() error {
	// Main request queue
	_, err := s.channel.QueueDeclare(
		"api.requests", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "dlx.exchange",
		}, // arguments
	)
	if err != nil {
		return err
	}

	// Response queue
	_, err = s.channel.QueueDeclare(
		"api.responses", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	// Dead Letter Queue
	_, err = s.channel.QueueDeclare(
		"api.dlq", // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	// DLX Exchange
	err = s.channel.ExchangeDeclare(
		"dlx.exchange", // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	// Bind DLQ to DLX
	err = s.channel.QueueBind(
		"api.dlq",      // queue name
		"api.dlq",      // routing key
		"dlx.exchange", // exchange
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	return nil
}

// ConsumeRequests потребляет сообщения из очереди запросов
func (s *ServiceMQ) ConsumeRequests(handler func(amqp.Delivery) error) error {
	if s.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	// Настраиваем QoS
	err := s.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := s.channel.Consume(
		"api.requests", // queue
		"",             // consumer
		false,          // auto-ack (false - ручное подтверждение)
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	// Запускаем обработчик в горутине
	go func() {
		for msg := range msgs {
			log.Printf("Received message: %s", msg.MessageId)

			// Обрабатываем сообщение
			if err := handler(msg); err != nil {
				log.Printf("Error processing message %s: %v", msg.MessageId, err)

				if msg.Redelivered {
					log.Printf("Message %s already redelivered, sending to DLQ", msg.MessageId)
					msg.Nack(false, false) // Не requeue, отправляем в DLQ
				} else {
					log.Printf("Message %s failed, requeuing", msg.MessageId)
					msg.Nack(false, true) // Requeue
				}
			} else {
				msg.Ack(false)
				log.Printf("Message %s processed successfully", msg.MessageId)
			}
		}
	}()

	log.Println("RabbitMQ consumer started for api.requests queue")
	return nil
}

// PublishResponse публикует ответ в очередь ответов или в ReplyTo очередь
func (s *ServiceMQ) PublishResponse(correlationID string, data interface{}, err error, replyTo string) error {
	if s.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	response := map[string]interface{}{
		"correlation_id": correlationID,
		"timestamp":      time.Now().UTC(),
	}

	if err != nil {
		response["status"] = "error"
		response["error"] = err.Error()
	} else {
		response["status"] = "success"
		response["data"] = data
	}

	body, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Определяем куда отправлять ответ
	targetQueue := "api.responses"
	if replyTo != "" {
		targetQueue = replyTo
	}

	publishing := amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		Body:          body,
		Timestamp:     time.Now(),
		DeliveryMode:  amqp.Persistent,
	}

	// Если это ответ в ReplyTo очередь, не делаем ее durable
	if replyTo != "" {
		publishing.DeliveryMode = amqp.Transient
	}

	err = s.channel.Publish(
		"",          // exchange
		targetQueue, // routing key
		false,       // mandatory
		false,       // immediate
		publishing,
	)
	if err != nil {
		return fmt.Errorf("failed to publish response to %s: %w", targetQueue, err)
	}

	log.Printf("Response published to %s for correlation_id: %s", targetQueue, correlationID)
	return nil
}
