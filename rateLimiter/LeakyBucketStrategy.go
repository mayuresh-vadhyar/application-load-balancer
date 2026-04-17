package rateLimiter

import (
	"strconv"
	"time"
)

type LeakyBucketStrategy struct {}

func (strategy *LeakyBucketStrategy) AllowRequest(rl RateLimiter, key string) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9
	capacity := float64(rl.limit)
	leakRate := capacity / rl.window.Seconds()
	data, err := rl.client.HGetAll(ctx, key).Result()

	if err != nil {
		return false, err
	}

	var tokens float64
	var last float64

	if len(data) == 0 {
		tokens = 0
		last = now
	} else {
		tokens, _ = strconv.ParseFloat(data["tokens"], 64)
		last, _ = strconv.ParseFloat(data["last"], 64)
	}

	elapsed := now - last
	tokens = tokens - (elapsed * leakRate)
	if tokens < 0 {
		tokens = 0
	}

	if tokens >= capacity {
		return false, nil
	}
	tokens += 1
	_, err = rl.client.HSet(ctx, key, map[string]interface{}{
		"tokens": tokens,
		"last":   now,
	}).Result()

	if err != nil {
		return false, err
	}

	rl.client.Expire(ctx, key, rl.window)
	return true, nil
}