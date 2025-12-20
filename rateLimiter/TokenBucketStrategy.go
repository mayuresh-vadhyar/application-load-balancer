package rateLimiter

import (
	"log"
	"os"
	"time"
)

var luaScript string

type TokenBucketStrategy struct {
	rate int
}

func (strategy TokenBucketStrategy) AllowRequest(rl RateLimiter, key string) (bool, error) {
	now := time.Now().UnixMilli()
	res := rl.client.Eval(ctx, luaScript, []string{key}, rl.limit, strategy.rate, now).String()
	return res == "ALLOW", nil
}

func (strategy TokenBucketStrategy) init() {
	data, err := os.ReadFile("rateLimiter/token_bucket.lua")
	if err != nil {
		log.Fatal(err)
	}
	luaScript = string(data)
}
