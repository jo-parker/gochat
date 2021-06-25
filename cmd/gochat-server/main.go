package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/go-redis/redis/v8"

	"github.com/jpparker/gochat/internal/pkg/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}
	port := os.Getenv("PORT")
	redisUrl := os.Getenv("REDIS_URL")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisUrl,
		Password: "",
		DB: 0,
	})

	server := &service.Server{
		Rdb: rdb,
	}

	http.HandleFunc("/websocket", server.HandleConnections)
	go server.HandleMessages()

	log.Println("[INFO] Web server starting at localhost:"+port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		log.Fatalln(err)
	}
}
