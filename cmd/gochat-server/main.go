package main

import (
	"log"
	"net/http"
	"os"
	"encoding/json"
	"context"

	"github.com/joho/godotenv"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"

	"github.com/jpparker/gochat/internal/pkg/model"
)

var (
	rdb *redis.Client
	clients = make(map[*websocket.Conn]bool)
	broadcaster = make(chan model.Message)
	upgrader = websocket.Upgrader {}
	ctx = context.Background()
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()
	clients[ws] = true

	if rdb.Exists(ctx, "chat_messages").Val() != 0 {
		sendPreviousMessages(ws)
	}

	for {
		var msg model.Message

		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}

		// Send new message to the channel
		broadcaster <- msg
	}
}

func storeInRedis(msg model.Message) {
	json, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	if err := rdb.RPush(ctx, "chat_messages", json).Err(); err != nil {
		log.Printf("Error: %v", err)
	}
}

func sendPreviousMessages(ws *websocket.Conn) {
	messages, err := rdb.LRange(ctx, "chat_messages", 0, -1).Result()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	for _, message := range messages {
		var msg model.Message
		json.Unmarshal([]byte(message), &msg)
		messageClient(ws, msg)
	}
}

func handleMessages() {
	for {
		// Grab any next message from channel
		msg := <- broadcaster

		storeInRedis(msg)
		messageClients(msg)
	}
}

func messageClient(client *websocket.Conn, msg model.Message) {
	err := client.WriteJSON(msg)
	if err != nil {
		log.Printf("Error: %v", err)

		client.Close()
		delete(clients, client)
	}
}

func messageClients(msg model.Message) {
	for client := range clients {
		messageClient(client, msg)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}
	port := os.Getenv("PORT")
	redisUrl := os.Getenv("REDIS_URL")

	rdb = redis.NewClient(&redis.Options{
		Addr: redisUrl,
		Password: "",
		DB: 0,
	})

	http.HandleFunc("/websocket", handleConnections)
	go handleMessages()

	log.Println("Web server starting at localhost:"+port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		log.Fatalln(err)
	}
}


