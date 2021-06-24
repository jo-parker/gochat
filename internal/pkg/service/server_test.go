package service

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"

	"github.com/jpparker/gochat/internal/pkg/model"
)

var (
	rdb *redis.Client
)

func TestStoreInRedis(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	rdb = redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		Password: "",
		DB: 0,
	})

	msg := model.Message {
		Username: "test",
		Text: "test message",
	}

	storeInRedis(msg, rdb)

	messages, _ := rdb.LRange(ctx, "chat_messages", 0, -1).Result()
	assert.Equal(t, len(messages), 1)
}
