package passport

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func TestCheckIPRateLimit(t *testing.T) {
	os.Setenv("REDIS_HOST", "localhost:6379")

	ip := net.IPv4bcast
	email := "johnsmith@example.com"

	l := &limiter{}

	//Set up
	err := l.Clear(context.Background(), email, ip)
	assert.Nil(t, err, "No errors should occur on del")

	results, TTL, remaining := l.CheckIP(context.Background(), ip)
	assert.Equal(t, results, true, "Initial IP limit should not be reached")
	assert.Equal(t, TTL, time.Duration(0), "TTL should be greater than 0")
	assert.Equal(t, remaining, ipMaxCount, "Remaining should still be %d", ipMaxCount)

	l.Inc(context.Background(), email, ip)

	results, TTL, remaining = l.CheckIP(context.Background(), ip)
	assert.Equal(t, results, true, "Initial IP limit should not be reached")
	assert.NotEqual(t, TTL, 0, "TTL should be greater than 0")
	assert.Equal(t, remaining, ipMaxCount-1, "Remaining should still be %d", ipMaxCount)
}

func TestCheckUserRateLimit(t *testing.T) {
	os.Setenv("REDIS_HOST", "localhost:6379")

	ip := net.IPv4bcast
	email := "johnsmith@example.com"

	l := &limiter{}

	//Set up
	err := l.Clear(context.Background(), email, ip)
	assert.Nil(t, err, "No errors should occur on del")

	results, TTL, remaining := l.CheckUser(context.Background(), email, ip)
	assert.Equal(t, results, true, "Initial IP limit should not be reached")
	assert.Equal(t, TTL, time.Duration(0), "TTL should be greater than 0")
	assert.Equal(t, remaining, userMaxCount, "Remaining should still be %d", userMaxCount)

	l.Inc(context.Background(), email, ip)

	results, TTL, remaining = l.CheckUser(context.Background(), email, ip)
	assert.Equal(t, results, true, "Initial IP limit should not be reached")
	assert.NotEqual(t, TTL, 0, "TTL should be greater than 0")
	assert.Equal(t, remaining, userMaxCount-1, "Remaining should still be %d", userMaxCount)

}
