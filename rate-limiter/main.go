package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter structure manages the rate limiting
type RateLimiter struct {
	mu          sync.Mutex
	limits      map[string]int
	lastRequest map[string]time.Time
	limit       int
	interval    time.Duration
}

func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		limits:      make(map[string]int),
		lastRequest: make(map[string]time.Time),
		limit:       limit,
		interval:    interval,
	}
}

// Allow checks whether a request is allowed or not
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Eğer bu IP için daha önce bir istek yapılmadıysa veya interval geçtiyse, sayaç sıfırlanır.
	if last, ok := rl.lastRequest[ip]; !ok || now.Sub(last) >= rl.interval {
		rl.limits[ip] = 0
		rl.lastRequest[ip] = now
	}

	// if limit not exceed allow
	if rl.limits[ip] < rl.limit {
		rl.limits[ip]++
		return true
	}

	return false
}

func RateLimitMiddleware(rl *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		// you can set user ip with header( for cloudflare ) like
		// forwarded := r.Header.Get("X-Forwarded-For")
		// fmt.Println(forwarded)

		if !rl.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// RateLimiter : 5 request / 3 sec
	rl := NewRateLimiter(5, 3*time.Second)

	// HTTP server created
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	})

	// Rate limiting middleware added
	http.ListenAndServe(":8080", RateLimitMiddleware(rl, mux))
}
