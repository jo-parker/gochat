package service

import (
	"log"
	"fmt"
	"os"
	"io"
	"time"
	"bufio"
	"encoding/json"

	"github.com/gorilla/websocket"

	"github.com/jpparker/gochat/internal/pkg/model"
)

var (
	output io.Writer = os.Stdout
)

type Client struct {
	Username	string
	Conn	*ConnShim
	Done	chan interface{}
	Text	chan string
}

type ConnShim struct {
	*websocket.Conn
}

func (c *Client) ReadUsernameInput(f *os.File) error {
	scanner := bufio.NewScanner(f)

	fmt.Fprint(output, "Enter username: ")

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			fmt.Fprintln(output, "Username cannot be empty.")
			fmt.Fprint(output, "Enter username: ")

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
		fmt.Fprintf(output, "%s %s: ", dt, c.Username)

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
		fmt.Fprintf(output, "\n%s %s: %s", dt, msg.Username, msg.Text)

		if output != os.Stdout {
			return
		}
	}

	close(c.Done)
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
