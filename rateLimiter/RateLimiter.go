package rateLimiter

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func InitializeRateLimiter(addr string, limit int, window time.Duration) *RateLimiter {
	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RateLimiter{
		client: redisClient,
		limit:  limit,
		window: window,
	}

}
