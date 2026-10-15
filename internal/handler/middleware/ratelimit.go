package middleware

import (
	"sync"
	"time"

	"dorm-issue-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	r        rate.Limit
	b        int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		r:        r,
		b:        b,
	}

	go i.cleanupVisitors()
	return i
}

func (i *IPRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := i.getVisitor(ip)

		if !limiter.Allow() {
			response.TooManyRequests(c, "มีการเรียกใช้งานมากเกินกำหนด (Rate limit exceeded) กรุณารอสักครู่")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (i *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	v, exists := i.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(i.r, i.b)
		i.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (i *IPRateLimiter) cleanupVisitors() {
	for {
		time.Sleep(10 * time.Minute)
		i.mu.Lock()
		for ip, v := range i.visitors {
			if time.Since(v.lastSeen) > 30*time.Minute {
				delete(i.visitors, ip)
			}
		}
		i.mu.Unlock()
	}
}
