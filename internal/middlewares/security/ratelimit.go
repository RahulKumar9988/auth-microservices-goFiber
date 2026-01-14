package security

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type Ratelimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *Ratelimiter {
	return &Ratelimiter{rdb: rdb}
}

func (r *Ratelimiter) Limit(key string, max int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		ip := c.IP()

		rediskey := fmt.Sprintf("rate:%s:%s", key, ip)

		count, err := r.rdb.Incr(ctx, rediskey).Result()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "rate limit error",
			})
		}

		if count == 1 {
			r.rdb.Expire(ctx, rediskey, window)
		}

		if count > int64(max) {
			return c.Status(429).JSON(fiber.Map{
				"error":       "too many request",
				"retry_after": window.Seconds(),
			})
		}

		return c.Next()
	}
}
