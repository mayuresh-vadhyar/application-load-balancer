package rateLimiter

import "github.com/mayuresh-vadhyar/application-load-balancer/constants"

type RateLimitStrategy interface {
	AllowRequest(rl RateLimiter, key string) (bool, error)
}

func GetRateLimitStrategy(algorithm string) RateLimitStrategy {
	if algorithm == "" {
		algorithm = constants.FIXED_WINDOW
	}
	switch algorithm {
	case constants.FIXED_WINDOW:
		return &FixedWindowStrategy{}
	}
	return nil
}
