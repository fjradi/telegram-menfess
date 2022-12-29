package adapter

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client, ctx: context.Background()}
}

func (r *Redis) SaveLatestSentMessageTimestamp(userId int, timestamp int64) error {
	return r.client.Set(r.ctx, strconv.Itoa(userId), timestamp, time.Duration(24*time.Hour)).Err()
}

func (r *Redis) GetLatestSentMessageTimestamp(userId int) (int64, error) {
	timestamp, err := r.client.Get(r.ctx, strconv.Itoa(userId)).Int64()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		log.Println("error getting latest sent message timestamp", err.Error())
		return 0, err
	}

	return timestamp, nil
}
