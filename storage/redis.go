package storage

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
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
	sub    *redis.PubSub
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

func (b *BackendRedis) PullLoop(dst chan Message) {
	b.sub = b.client.Subscribe()
	defer b.sub.Close()
	var err error
	var msgi interface{}
	var msg Message

	for {
		msgi, err = b.sub.Receive()
		if err != nil {
			logrus.Errorf("recv error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		switch mi := msgi.(type) {
		case *redis.Subscription:
			logrus.Infof("subscribe succeeded to %s", mi.Channel)
		case *redis.Message:
			err = json.Unmarshal([]byte(mi.Payload), &msg)
			if err != nil {
				logrus.Errorf("msg error: %v", err)
				continue
			}
			log.Debugf("recv msg: %v", msg)
		case *redis.Pong:
			// pong received
		default:
			// handle error
		}
	}
}

func (b *BackendRedis) Subscribe(channel string) error {
	return b.sub.Subscribe(channel)
}
