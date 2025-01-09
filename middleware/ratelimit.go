package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  1,
	}

	store := memory.NewStore()

	limiterInstance := limiter.New(store, rate)

	return func(c *gin.Context) {
		key := c.ClientIP()

		context, err := limiterInstance.Get(c, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			c.JSON(429, gin.H{
				"code":    "1",
				"message": "Too many requests. Try again later.",
			})
			return
		}

		c.Next()
	}
}
