package rateLimiter

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/mayuresh-vadhyar/application-load-balancer/Redis"
	"github.com/mayuresh-vadhyar/application-load-balancer/config"
	"github.com/redis/go-redis/v9"
)

type Config = config.Config
type Identifier string

type RateLimiter struct {
	client     *redis.Client
	strategy   RateLimitStrategy
	identifier Identifier
	limit      int
	window     time.Duration
}

const (
	IP           Identifier = "IP"
	ApiKey       Identifier = "ApiKey"
	UserID       Identifier = "UserID"
	ApiPath      Identifier = "ApiPath"
	Resource     Identifier = "Resource"
	defaultValue string     = "global"
)

var ctx = context.Background()
var prefix = "_ratelimiter"

func isValidIdentifier(id Identifier) bool {
	identifiers := []Identifier{IP, ApiKey, UserID, ApiPath, Resource}
	return slices.Contains(identifiers, id)
}

func getIdentifierStrategy(config RateLimitConfig) Identifier {
	identifier := Identifier(config.Identifier)
	if config.Identifier == "" || !isValidIdentifier(identifier) {
		return IP
	}

	return identifier
}

func (rl RateLimiter) getIdentifier(r *http.Request) string {
	switch rl.identifier {
	case IP:
		return r.RemoteAddr
	case ApiKey:
		return r.Header.Get("X-API-Key")
	case UserID:
		return r.Header.Get("X-User-ID")
	case ApiPath:
		return r.URL.Path
	case Resource:
		return strings.Split(r.URL.Path, "/")[1]
	}
	return defaultValue
}

func GetRateLimiter() *RateLimiter {
	config := config.GetConfig()
	if config.RedisURL == "" || !config.RateLimit.Enable {
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
	strategy := GetRateLimitStrategy(config.RateLimit)
	identifierStrategy := getIdentifierStrategy(config.RateLimit)

	client := Redis.GetClient()

	if client == nil {
		return nil
	}

	return &RateLimiter{
		client:     client,
		strategy:   strategy,
		identifier: identifierStrategy,
		limit:      limit,
		window:     window,
	}

}

func keyGenerator(token string) string {
	return fmt.Sprintf("%s:%s", prefix, token)
}

func (rl RateLimiter) allowRequest(key string) (bool, error) {
	return rl.strategy.AllowRequest(rl, key)
}

func (rl RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := rl.getIdentifier(r)
		key := keyGenerator(token)
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
