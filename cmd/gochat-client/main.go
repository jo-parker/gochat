package main

import (
	"log"
	"fmt"
	"os"
	"os/signal"
	"time"
	"bufio"
	"encoding/json"

	"github.com/joho/godotenv"
	"github.com/gorilla/websocket"
	"github.com/jpparker/gochat/internal/pkg/model"
)

var (
	username string
	done = make(chan interface{})
	interrupt = make(chan os.Signal)
	text = make(chan string)
	scanner = bufio.NewScanner(os.Stdin)
)

func readUsernameInput() {
	fmt.Print("Enter username: ")

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			fmt.Println("Username cannot be empty.")
			fmt.Print("Enter username: ")

			continue
		}

		username = text
		break
	}
}

func readMessageInput() {
	for scanner.Scan() {
		dt := time.Now().Format("2006/01/02 15:04:05")
		fmt.Printf("%s %s: ", dt, username)

		out := scanner.Text()

		if out != "" {
			text <- out
		}
	}
}

func sendMessage(conn *websocket.Conn, text string) error {
	msg := model.Message{
		Username: username,
		Text: text,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return err
	}

	return nil
}

func receiveHandler(conn *websocket.Conn) {
	defer close(done)

	for {
		_, data, err := conn.ReadMessage()

		if err != nil {
			log.Println("Error in receive:", err)
			return
		}

		var msg model.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println(err)
		}

		dt := time.Now().Format("2006/01/02 15:04:05")
		fmt.Printf("\n%s %s: %s", dt, msg.Username, msg.Text)
	}
}

func main () {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	readUsernameInput()

	signal.Notify(interrupt, os.Interrupt)
	socketUrl := os.Getenv("SOCKET_URL")

	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatalln("Error connecting to GoChat server: ", err)
	}
	defer conn.Close()

	go receiveHandler(conn)
	go readMessageInput()

	for {
		select {
		case msg := <- text:
			// New message from STDIN. Write to WebSocket
			err := sendMessage(conn, msg)

			if err != nil {
				log.Println("Error sending message to WebSocket: ", err)
			}
		case <- interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our WebSocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
					log.Println("Error during closing websocket: ", err)
					return
			}

			select {
			case <- done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <- time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}

	if scanner.Err() != nil {
		log.Println("Error: ", scanner.Err())
	}
}