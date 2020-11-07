package redisrate

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

var pool *redis.Pool

func newPool(addr string, password string) *redis.Pool {
	return &redis.Pool{
		Wait:        true,
		MaxIdle:     3,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		// Dial or DialContext must be set.
		// When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr,
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second))
			if err != nil {
				return nil, err
			}

			if len(password) != 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}

			if _, err := c.Do("SELECT", 0); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}

func rateLimiter() *Limiter {
	pool = newPool("127.0.0.1:6379", "")
	return NewLimiter(pool)
}

func TestAllow(t *testing.T) {
	var l = rateLimiter()
	limit := PerSecond(10)

	res, err := l.Allow("test_id", limit)
	assert.Nil(t, err)
	assert.True(t, res.Allowed)
	assert.Equal(t, res.Remaining, 9)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 100*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 2)
	assert.Nil(t, err)
	assert.True(t, res.Allowed)
	assert.Equal(t, res.Remaining, 7)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 300*time.Millisecond, float64(50*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 1000)
	assert.Nil(t, err)
	assert.False(t, res.Allowed)
	assert.Equal(t, res.Remaining, 0)
	assert.InDelta(t, res.RetryAfter, 99*time.Second, float64(time.Second))
	assert.InDelta(t, res.ResetAfter, 300*time.Millisecond, float64(50*time.Millisecond))
}

func BenchmarkAllow(b *testing.B) {
	var l = rateLimiter()
	limit := PerSecond(100)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := l.Allow("foo", limit)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
