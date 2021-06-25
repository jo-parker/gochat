package service

import (
	"log"
	"net/http"
	"encoding/json"
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"

	"github.com/jpparker/gochat/internal/pkg/model"
)

type Server struct {
	Rdb *redis.Client
}

var (
	clients = make(map[*websocket.Conn]bool)
	broadcaster = make(chan model.Message)
	upgrader = websocket.Upgrader{}
	ctx = context.Background()
)

func (s *Server) HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()
	clients[ws] = true

	if s.Rdb.Exists(ctx, "chat_messages").Val() != 0 {
		sendPreviousMessages(ws, s.Rdb)
	}

	for {
		var msg model.Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}

		broadcaster <- msg
	}
}

func (s *Server) HandleMessages() {
	for {
		msg := <- broadcaster

		storeInRedis(msg, s.Rdb)
		messageClients(msg)
	}
}

func storeInRedis(msg model.Message, rdb *redis.Client) {
	json, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	if err := rdb.RPush(ctx, "chat_messages", json).Err(); err != nil {
		log.Printf("[ERROR] %v", err)
	}
}

func sendPreviousMessages(ws *websocket.Conn, rdb *redis.Client) {
	messages, err := rdb.LRange(ctx, "chat_messages", 0, -1).Result()
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	for _, message := range messages {
		var msg model.Message
		json.Unmarshal([]byte(message), &msg)
		messageClient(ws, msg)
	}
}

func messageClient(client *websocket.Conn, msg model.Message) {
	err := client.WriteJSON(msg)
	if err != nil {
		log.Printf("[ERROR] %v", err)

		client.Close()
		delete(clients, client)
	}
}

func messageClients(msg model.Message) {
	for client := range clients {
		messageClient(client, msg)
	}
}
