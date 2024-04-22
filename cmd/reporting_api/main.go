package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func main() {
	r := gin.Default()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	r.GET("/message/list", func(c *gin.Context) {
		sender := c.Query("sender")
		receiver := c.Query("receiver")

		if sender == "" || receiver == "" {
			c.JSON(400, gin.H{"error": "Sender and receiver parameters are required"})
			return
		}

		key := sender + "_" + receiver
		exists, err := rdb.Exists(key).Result()
		if err != nil {
			log.Fatalf("Failed to check if key exists: %v", err)
		}

		array := []string{}

		if exists != 0 {
			data, err := rdb.Get(key).Bytes()
			if err != nil {
				log.Fatalf("Failed to get value from Redis: %v", err)
			}

			err = json.Unmarshal(data, &array)
			if err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}
		}

		c.JSON(http.StatusOK, array)
	})

	r.Run()
}
