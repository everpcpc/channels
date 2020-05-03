package storage

import (
	"encoding/json"

	"github.com/go-redis/redis/v7"
)

func NewRedisBackend(addr string) (*BackendRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	b := &BackendRedis{
		client: client,
	}
	return b, nil
}

type BackendRedis struct {
	client *redis.Client
}

func (b *BackendRedis) Save(msg Message) (err error) {
	var s []byte
	s, err = json.Marshal(msg)
	if err != nil {
		return
	}
	_, err = b.client.Publish(msg.Channel, s).Result()
	return
}
