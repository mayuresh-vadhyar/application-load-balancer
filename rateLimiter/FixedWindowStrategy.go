package rateLimiter

import "log"

type FixedWindowStrategy struct {
}

func (strategy FixedWindowStrategy) AllowRequest(rl RateLimiter, key string) (bool, error) {
	// TODO: Apply lock
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		log.Println(err)
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
