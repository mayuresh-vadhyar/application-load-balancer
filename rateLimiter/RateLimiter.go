package rateLimiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mayuresh-vadhyar/application-load-balancer/config"
	"github.com/redis/go-redis/v9"
)

type Config = config.Config

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

var ctx = context.Background()
var prefix = "_ratelimiter"

func InitializeRateLimiter() *RateLimiter {
	config := config.GetConfig()
	if config.RedisURL == "" {
		return nil
	}
	window, err := time.ParseDuration(config.RateLimit.Window)
	if err != nil {
		window = (time.Minute * 2)
	}

	limit := config.RateLimit.Limit
	if limit == 0 {
		limit = 25
	}

	client := redis.NewClient(&redis.Options{
		Addr: config.RedisURL,
	})
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}

}

func keyGenerator(token string) string {
	return fmt.Sprintf("%s:%s", prefix, token)
}

func (rl RateLimiter) allowRequest(key string) (bool, error) {
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		_, err = rl.client.Expire(ctx, key, rl.window).Result()
		if err != nil {
			return false, err
		}
	}

	if count > int64(rl.limit) {
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
		next.ServeHTTP(w, r)
	})
}
