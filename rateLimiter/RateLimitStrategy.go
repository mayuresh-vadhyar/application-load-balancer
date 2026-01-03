package rateLimiter

import (
	"log"

	"github.com/mayuresh-vadhyar/application-load-balancer/config"
)

type RateLimitConfig = config.RateLimitConfig
type RateLimitStrategy interface {
	AllowRequest(rl RateLimiter, key string) (bool, error)
}

const (
	FIXED_WINDOW = "FW"
	TOKEN_BUCKET = "TB"
)

func GetRateLimitStrategy(config RateLimitConfig) RateLimitStrategy {
	algorithm := config.Strategy
	if algorithm == "" {
		algorithm = FIXED_WINDOW
	}

	switch algorithm {
	case FIXED_WINDOW:
		return &FixedWindowStrategy{}
	case TOKEN_BUCKET:
		if config.Rate <= 0 {
			log.Fatalf("Token bucket refill rate missing")
		}
		strategy := &TokenBucketStrategy{rate: config.Rate}
		strategy.init()
		return strategy
	}
	return nil
}
