package rateLimiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

var ctx = context.Background()
var prefix = "_ratelimiter"

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

func keyGenerator(token string) string {
	return fmt.Sprintf("%s:%s", prefix, token)
}

func (rl RateLimiter) allowRequest(key string) (bool, error) {
	pipe := rl.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	if incr.Val() > int64(rl.limit) {
		return false, nil
	}

	return true, nil
}

func (rl RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		key := keyGenerator(ip)
		allowed, err := rl.allowRequest(key)

		if err != nil {
			http.Error(w, "Rate Limit Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}
	})
}
