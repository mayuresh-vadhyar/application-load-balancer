package rateLimiter

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowStrategy struct {}

func (strategy SlidingWindowStrategy) AllowRequest(rl RateLimiter, key string) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - rl.window.Nanoseconds()
	pipe := rl.client.TxPipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// TODO: replace with Member: uuid.New().String()
	pipe.ZAdd(ctx,key,redis.Z{
		Score: float64(now),
		Member: now,
	})
	count := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	if count.Val() > int64(rl.limit) {
		return false, nil
	}

	return  true, nil
}