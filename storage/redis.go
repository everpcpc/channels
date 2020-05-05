package storage

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

type RedisConfig struct {
	Network  string
	Addr     string
	Password string
	DB       int
}

func NewRedisBackend(cfg *RedisConfig) (*BackendRedis, error) {
	client := redis.NewClient(&redis.Options{
		Network:  cfg.Network,
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
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
	_, err = b.client.Publish(msg.To, s).Result()
	return
}

func (b *BackendRedis) PullLoop(dst chan Message) {
	// default subscribe to #announce
	b.sub = b.client.Subscribe("#announce")
	defer b.sub.Close()
	var err error
	var msgi interface{}

	for {
		msgi, err = b.sub.Receive()
		if err != nil {
			logrus.Errorf("recv error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		switch mi := msgi.(type) {
		case *redis.Subscription:
			logrus.Infof("%s %s", mi.Kind, mi.Channel)
		case *redis.Message:
			var msg Message
			err = json.Unmarshal([]byte(mi.Payload), &msg)
			if err != nil {
				logrus.Errorf("msg error: %v", err)
				continue
			}
			logrus.Debugf("recv msg: %v", msg)
			dst <- msg
		case *redis.Pong:
			// pong received
		default:
			logrus.Warnf("recv unknown msg: %v", mi)
		}
	}
}

func (b *BackendRedis) Subscribe(channel string) error {
	return b.sub.Subscribe(channel)
}

func (b *BackendRedis) UnSubscribe(channel string) error {
	return b.sub.Unsubscribe(channel)
}

func (b *BackendRedis) GetToken(token string) (data *TokenData, err error) {
	var res string
	res, err = b.client.Get("token:" + token).Result()
	if err != nil {
		return
	}
	data = new(TokenData)
	err = json.Unmarshal([]byte(res), data)
	return
}

func (b *BackendRedis) AddToken(token string, data *TokenData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	pipe := b.client.Pipeline()
	pipe.SAdd("tokens", token)
	pipe.Set("token:"+token, jsonData, 0)
	_, err = pipe.Exec()
	return err
}

func (b *BackendRedis) DeleteToken(token string) error {
	pipe := b.client.Pipeline()
	pipe.SRem("tokens", token)
	pipe.Del("token:" + token)
	_, err := pipe.Exec()
	return err
}
