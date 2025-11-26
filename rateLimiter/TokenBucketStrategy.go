package rateLimiter

import "time"

// go:embed token_bucket.lua
var luaScript string

type TokenBucketStrategy struct {
	rate int
}

func (strategy TokenBucketStrategy) AllowRequest(rl RateLimiter, key string) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9

	res, err := rl.client.Eval(ctx, luaScript, []string{key}, rl.limit, 5, now).Int()

	if err != nil {
		return false, err
	}

	return res == 1, nil
}