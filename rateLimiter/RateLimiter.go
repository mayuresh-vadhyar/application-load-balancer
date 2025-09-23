package rateLimiter

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

var prefix = "_ratelimiter:"

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

func keyGenerator(value string) string {}

func (rl RateLimiter) allowRequest() (bool, error) {}

func (rl RateLimiter) RateLimit(next http.Handler) http.Handler {}
