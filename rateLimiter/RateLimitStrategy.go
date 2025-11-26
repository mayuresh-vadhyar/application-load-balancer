package rateLimiter

import (
	"github.com/mayuresh-vadhyar/application-load-balancer/config"
	"github.com/mayuresh-vadhyar/application-load-balancer/constants"
)

type RateLimitConfig = config.RateLimitConfig
type RateLimitStrategy interface {
	AllowRequest(rl RateLimiter, key string) (bool, error)
}

func GetRateLimitStrategy(config RateLimitConfig) RateLimitStrategy {
	algorithm := config.Strategy
	if algorithm == "" {
		algorithm = constants.FIXED_WINDOW
	}

	switch algorithm {
	case constants.FIXED_WINDOW:
		return &FixedWindowStrategy{}
	case constants.TOKEN_BUCKET:
		return &TokenBucketStrategy{rate: config.Rate}
	}
	return nil
}
