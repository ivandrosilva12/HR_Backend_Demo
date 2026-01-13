package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RateLimiterRedisMiddleware aplica rate limiting por IP com Redis como backend
func RateLimiterRedisMiddleware(limit int64, period time.Duration, client *redis.Client) gin.HandlerFunc {
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	store, err := redisstore.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "ratelimiter",
		MaxRetry: 3,
	})
	if err != nil {
		panic("Erro ao criar Redis RateLimiter Store: " + err.Error())
	}

	return ginlimiter.NewMiddleware(limiter.New(store, rate))
}
