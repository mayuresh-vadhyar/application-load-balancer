package rateLimiter

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

var slidingWindowLuaScript string

type SlidingWindowStrategy struct{}

func (strategy SlidingWindowStrategy) AllowRequest(rl RateLimiter, key string) (bool, error) {
	now := time.Now().UnixMilli()
	member := uuid.New().String()

	res, err := rl.client.Eval(ctx, slidingWindowLuaScript, []string{key}, rl.limit, rl.window.Milliseconds(), now, member).Result()
	if err != nil {
		return false, err
	}

	return res == "ALLOW", nil
}

func (strategy SlidingWindowStrategy) init() {
	data, err := os.ReadFile("rateLimiter/sliding_window.lua")
	if err != nil {
		log.Fatal(err)
	}
	slidingWindowLuaScript = string(data)
}
