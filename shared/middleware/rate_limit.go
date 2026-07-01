package middleware

import (
	"net"
	"sync"
	"time"

	"github.com/bastion-framework/bast"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	r        rate.Limit
	b        int
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	l := &ipLimiter{
		visitors: make(map[string]*visitor),
		r:        rate.Limit(rps),
		b:        burst,
	}
	go l.cleanup()
	return l
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[ip]
	if !ok {
		lim := rate.NewLimiter(l.r, l.b)
		l.visitors[ip] = &visitor{limiter: lim, lastSeen: time.Now()}
		return lim
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanup removes idle entries every 5 minutes.
func (l *ipLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

// RateLimit returns a per-IP token bucket middleware.
// rps is the sustained rate; burst allows short spikes above that rate.
func RateLimit(rps float64, burst int) bast.MiddlewareFunc {
	l := newIPLimiter(rps, burst)
	return func(next bast.HandlerFunc) bast.HandlerFunc {
		return func(ctx *bast.Ctx) bast.Response {
			ip, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
			if err != nil {
				ip = ctx.Request.RemoteAddr
			}
			if !l.get(ip).Allow() {
				return ctx.Error(bast.ErrTooManyRequests("rate limit exceeded, try again later"))
			}
			return next(ctx)
		}
	}
}
