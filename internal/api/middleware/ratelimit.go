package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// limiterStore holds a rate limiter per IP per route
type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newStore(r rate.Limit, b int) *limiterStore {
	return &limiterStore{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (s *limiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock() // make sure to use mutexes when there is potential of concurrent writes in maps. can corrupt data.
	defer s.mu.Unlock()

	if lim, exists := s.limiters[ip]; exists {
		return lim
	}
	lim := rate.NewLimiter(s.r, s.b)
	s.limiters[ip] = lim
	return lim
}

// RateLimit returns a gin middleware with the given rate and burst
func RateLimit(r rate.Limit, burst int) gin.HandlerFunc {
	store := newStore(r, burst)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		lim := store.get(ip)
		if !lim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}
