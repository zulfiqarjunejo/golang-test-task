package main

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Check Redis connection
	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

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

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	go func() {
		for m := range msgs {
			var msg Message
			err := json.Unmarshal(m.Body, &msg)
			if err != nil {
				log.Printf("Failed to unmarshal JSON: %v", err)
				return
			}

			key := msg.Sender + "_" + msg.Receiver
			exists, err := rdb.Exists(key).Result()
			if err != nil {
				log.Fatalf("Failed to check if key exists: %v", err)
			}

			if exists == 0 {
				array := []string{msg.Message}

				// Convert array to JSON
				data, err := json.Marshal(array)
				if err != nil {
					log.Fatalf("Failed to marshal JSON: %v", err)
				}

				// Insert array into Redis
				err = rdb.Set(key, data, 0).Err()
				if err != nil {
					log.Fatalf("Failed to insert array into Redis: %v", err)
				}

				log.Println("Array inserted into Redis")
			} else {
				data, err := rdb.Get(key).Bytes()
				if err != nil {
					log.Fatalf("Failed to get value from Redis: %v", err)
				}

				var array []string
				err = json.Unmarshal(data, &array)
				if err != nil {
					log.Fatalf("Failed to unmarshal JSON: %v", err)
				}

				// Add new message to the beginning of the array
				array = append([]string{msg.Message}, array...)

				// Marshal updated array into JSON
				newData, err := json.Marshal(array)
				if err != nil {
					log.Fatalf("Failed to marshal JSON: %v", err)
				}

				// Set updated array back to Redis
				err = rdb.Set(key, newData, 0).Err()
				if err != nil {
					log.Fatalf("Failed to set value in Redis: %v", err)
				}

				log.Println("New message added to the array in Redis")
			}
		}
	}()

	log.Println("Waiting for messages. To exit, press Ctrl+C")
	select {}
}
