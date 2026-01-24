package Redis

import (
	"context"
	"sync"

	"github.com/mayuresh-vadhyar/application-load-balancer/config"
	"github.com/redis/go-redis/v9"
)

var initOnce sync.Once
var client *redis.Client

func GetClient() *redis.Client {
	initOnce.Do(func() {
		config := config.GetConfig()
		ctx := context.Background()
		client = redis.NewClient(&redis.Options{
			Addr: config.RedisURL,
		})

		_, pingErr := client.Ping(ctx).Result()
		if pingErr != nil {
			client = nil
		}
	})
	return client
}
