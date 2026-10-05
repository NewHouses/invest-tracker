package web

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	loginFailureLimit  = 5
	loginLockDuration  = time.Minute
	loginFailureWindow = 15 * time.Minute
	loginLimiterMaxIPs = 1000
)

type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]*loginLimitEntry
}

type loginLimitEntry struct {
	failures    int
	lastFailure time.Time
	lockedUntil time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{entries: map[string]*loginLimitEntry{}}
}

func (l *loginLimiter) allow(r *http.Request, now time.Time) (bool, time.Duration) {
	ip := clientIP(r)
	if ip == "" {
		return true, 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	e := l.entries[ip]
	if e == nil {
		return true, 0
	}
	if now.Sub(e.lastFailure) > loginFailureWindow && !now.Before(e.lockedUntil) {
		delete(l.entries, ip)
		return true, 0
	}
	if now.Before(e.lockedUntil) {
		return false, e.lockedUntil.Sub(now)
	}
	return true, 0
}

func (l *loginLimiter) failure(r *http.Request, now time.Time) {
	ip := clientIP(r)
	if ip == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	e := l.entries[ip]
	if e == nil {
		e = &loginLimitEntry{}
		l.entries[ip] = e
	}
	if now.Sub(e.lastFailure) > loginFailureWindow {
		e.failures = 0
	}
	e.failures++
	e.lastFailure = now
	if e.failures >= loginFailureLimit {
		e.lockedUntil = now.Add(loginLockDuration)
	}
	l.evictLocked(now)
}

func (l *loginLimiter) success(r *http.Request) {
	ip := clientIP(r)
	if ip == "" {
		return
	}
	l.mu.Lock()
	delete(l.entries, ip)
	l.mu.Unlock()
}

func (l *loginLimiter) evictLocked(now time.Time) {
	if len(l.entries) <= loginLimiterMaxIPs {
		return
	}
	for ip, e := range l.entries {
		if len(l.entries) <= loginLimiterMaxIPs {
			return
		}
		if now.Sub(e.lastFailure) > loginFailureWindow && !now.Before(e.lockedUntil) {
			delete(l.entries, ip)
		}
	}
	for ip := range l.entries {
		if len(l.entries) <= loginLimiterMaxIPs {
			return
		}
		delete(l.entries, ip)
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func retryAfterSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int((d + time.Second - time.Nanosecond) / time.Second)
}
