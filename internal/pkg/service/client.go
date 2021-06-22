package service

import (
	"log"
	"fmt"
	"os"
	"time"
	"bufio"
	"encoding/json"

	"github.com/gorilla/websocket"

	"github.com/jpparker/gochat/internal/pkg/model"
)

type Client struct {
	Username	string
	Conn	*websocket.Conn
	Done	chan interface{}
	Text	chan string
}

func (c *Client) ReadUsernameInput(f *os.File) error {
	scanner := bufio.NewScanner(f)

	fmt.Print("Enter username: ")

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			fmt.Println("Username cannot be empty.")
			fmt.Print("Enter username: ")

			continue
		}

		c.Username = text
		break
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error: ", err)
		return err
	}

	return nil
}

func (c *Client) ReadMessageInput(f *os.File) error {
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		dt := time.Now().Format("2006/01/02 15:04:05")
		fmt.Printf("%s %s: ", dt, c.Username)

		out := scanner.Text()

		if out != "" {
			c.Text <- out
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error: ", err)
		return err
	}

	return nil
}

func (c *Client) ReceiveHandler() {
	defer close(c.Done)

	for {
		_, data, err := c.Conn.ReadMessage()

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

func (c *Client) SendMessage(text string) error {
	msg := model.Message{
		Username: c.Username,
		Text: text,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return err
	}

	return nil
}
