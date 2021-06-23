package service

import (
	"testing"
	"os"
	"io"
	"fmt"
	"io/ioutil"
	"log"
	"bytes"
	"strings"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/gorilla/websocket"
)

var (
	message string
	objectString = `{"Username":"test","Text":"test message"}`
	client = &Client{
		Text: make(chan string, 1),
		Conn: &ConnShim{},
	}
)

func (conn *ConnShim) ReadMessage() (messageType int, p []byte, err error) {
	var r io.Reader
	r, err = writeTempFile(message)
	if err != nil {
		return 0, nil, err
	}

	p, err = ioutil.ReadAll(r)
	return websocket.TextMessage, p, err
}

func (conn *ConnShim) WriteMessage(messageType int, p []byte) error {
	fmt.Fprintf(output, "message type: %d, text: %s", websocket.TextMessage, string(p))
	return nil
}

func TestReadUsernameInputIsEmpty(t *testing.T) {
	tmpfile, err := writeTempFile("")
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	output = buf

	if err := client.ReadUsernameInput(tmpfile); err != nil {
		t.Errorf("ReadUsernameInput failed: %v", err)
	}

	assert.Equal(t, client.Username, "")
}

func TestReadUsernameInput(t *testing.T) {
	tmpfile, err := writeTempFile("username")
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	output = buf

	if err := client.ReadUsernameInput(tmpfile); err != nil {
		t.Errorf("ReadUsernameInput failed: %v", err)
	}

	assert.Equal(t, client.Username, "username")
}

func TestReadMessageInputIsEmpty(t *testing.T) {
	tmpfile, err := writeTempFile("")
	if err != nil {
		log.Fatal(err)
	}

	if err := client.ReadMessageInput(tmpfile); err != nil {
		t.Errorf("ReadMessageInput failed: %v", err)
	}

	assert.Equal(t, len(client.Text), 0)
}

func TestReadMessageInput(t *testing.T) {
	tmpfile, err := writeTempFile("test message")
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	output = buf

	if err := client.ReadMessageInput(tmpfile); err != nil {
		t.Errorf("ReadMessageInput failed: %v", err)
	}

	assert.Equal(t, len(client.Text), 1)
}

func TestReceiveHandler(t *testing.T) {
	gostub.Stub(&message, objectString)

	buf := &bytes.Buffer{}
	output = buf

	client.ReceiveHandler()

	assert.True(t, strings.Contains(buf.String(), "test: test message"))
}

func TestSendMessage(t *testing.T) {
	client.Username = "test"

	buf := &bytes.Buffer{}
	output = buf

	client.SendMessage("test message")

	assert.True(t, strings.Contains(buf.String(), "message type: 1, text: " + objectString))
}

func writeTempFile(c string) (*os.File, error) {
	content := []byte(c)

	tmpfile, err := ioutil.TempFile("", "test_tmpfile")
	if err != nil {
		return nil, err
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		return nil, err
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		return nil, err
	}

	return tmpfile, nil
}
