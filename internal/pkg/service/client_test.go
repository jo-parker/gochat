package service

import (
	"testing"
	"os"
	"io/ioutil"
	"log"

	"github.com/stretchr/testify/assert"
)

var client = &Client{
	Text: make(chan string),
}

func TestReadUsernameInputIsEmpty(t *testing.T) {
	tmpfile, err := writeTempFile("")
	if err != nil {
		log.Fatal(err)
	}

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

	if err := client.ReadUsernameInput(tmpfile); err != nil {
		t.Errorf("ReadUsernameInput failed: %v", err)
	}

	assert.Equal(t, client.Username, "username")
}

/*
func TestReadMessageInput(t *testing.T) {
	tmpfile, err := writeTempFile("test message")
	if err != nil {
		log.Fatal(err)
	}

	if err := client.ReadMessageInput(tmpfile); err != nil {
		t.Errorf("ReadMessageInput failed: %v", err)
	}

	assert.Equal(t, len(client.Text), 1)
}
*/

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
