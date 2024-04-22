package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func main() {
	r := gin.Default()

	r.POST("/message", func(c *gin.Context) {
		var message Message
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
		if err != nil {
			log.Fatalf("Failed to connect to RabbitMQ: %v", err)
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			log.Fatalf("Failed to open a channel: %v", err)
		}
		defer ch.Close()

		// Ensure that the queue exists. i-e: create the queue if it does not exist.
		queue, err := ch.QueueDeclare(
			"message_queue", // name
			false,           // durable
			false,           // delete when unused
			false,           // exclusive
			false,           // no-wait
			nil,             // arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare a queue: %v", err)
		}

		// Create a JSON message
		body, err := json.Marshal(message)
		if err != nil {
			log.Fatalf("Failed to encode message to JSON: %v", err)
		}

		// Publish message to the queue
		err = ch.Publish(
			"",         // exchange
			queue.Name, // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			log.Fatalf("Failed to publish message: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Message published successfully"})
	})

	r.Run()
}
