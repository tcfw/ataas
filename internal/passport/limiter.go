package passport

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	passportAPI "pm.tcfw.com.au/source/ataas/api/pb/passport"
	"pm.tcfw.com.au/source/ataas/internal/broadcast"
)

const (
	userMaxCount = 3
	userLimitTTL = time.Minute * 5
	ipLimitTTL   = time.Minute * 60
	ipMaxCount   = 100
)

type limiter struct {
	log         *logrus.Logger
	redisClient *redis.Client
}

func (l *limiter) Clear(ctx context.Context, email string, ip net.IP) error {
	client, err := l.cache(ctx)
	if err != nil {
		return err
	}

	userKey := l.userCountKey(email, ip)
	ipKey := l.ipCountKey(ip)

	count, err := client.Del(userKey, ipKey).Result()
	if err != nil || count != 2 {
		return err
	}

	return nil
}

type keyTTL struct {
	key string
	ttl time.Duration
}

func (l *limiter) Inc(ctx context.Context, email string, ip net.IP) {
	keysToIncrement := []keyTTL{
		{key: l.userCountKey(email, ip), ttl: userLimitTTL},
		{key: l.ipCountKey(ip), ttl: ipLimitTTL},
	}

	client, err := l.cache(ctx)
	if err != nil {
		return
	}
	defer client.Close()

	for _, limit := range keysToIncrement {
		exists := len(client.Keys(limit.key).Val())
		if exists == 0 {
			client.Set(limit.key, 0, limit.ttl)
		}
		client.Incr(limit.key)
	}
}

func (l *limiter) CheckIP(ctx context.Context, ip net.IP) (bool, time.Duration, int) {
	key := l.ipCountKey(ip)

	return l.checkKeyRateLimit(ctx, key, ipMaxCount)
}

func (l *limiter) CheckUser(ctx context.Context, username string, ip net.IP) (bool, time.Duration, int) {
	key := l.userCountKey(username, ip)

	return l.checkKeyRateLimit(ctx, key, userMaxCount)
}

func (l *limiter) checkKeyRateLimit(ctx context.Context, key string, limit int) (bool, time.Duration, int) {
	client, err := l.cache(ctx)
	if err != nil {
		return true, 0, limit //Allow on error
	}
	defer client.Close()

	exists := len(client.Keys(key).Val())

	if exists == 0 {
		return true, 0, limit //Allow if count doesn't exist
	}

	count, err := client.Get(key).Int64()
	if err != nil {
		return true, 0, limit //Allow on error
	}

	ttl := client.TTL(key).Val()
	if int(count) <= limit {
		return true, ttl, limit - int(count)
	}

	return false, ttl, limit
}

func (l *limiter) userCountKey(email string, ip net.IP) string {
	return "login_rate:" + ip.String() + ":" + email
}

func (l *limiter) ipCountKey(ip net.IP) string {
	return "login_rate:" + ip.String()
}

func (l *limiter) cache(ctx context.Context) (*redis.Client, error) {
	if l.redisClient != nil {
		return l.redisClient, nil
	}

	redisConnHost, exists := os.LookupEnv("REDIS_HOST")
	if !exists {
		redisConnHost = "redis:6379"
	}

	l.redisClient = redis.NewClient(&redis.Options{
		Addr:     redisConnHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := l.redisClient.Ping().Result(); err != nil {
		return nil, err
	}

	return l.redisClient, nil
}

func (l *limiter) ReachedResp(ctx context.Context, remoteIP net.IP, ttl time.Duration) (*passportAPI.AuthResponse, error) {
	b, err := broadcast.Driver()
	if err != nil {
		return nil, err
	}

	grpc.SendHeader(ctx, metadata.Pairs("Grpc-Metadata-X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix())))
	b.Publish("passport", broadcast.AuthenticateEvent{Event: &broadcast.Event{Type: "vanga.passport.limit_reached"}, Err: "limit reached: ip", IP: remoteIP.String()})
	return nil, status.Errorf(codes.ResourceExhausted, "IP rate limit exceeded. Wait %s before making another request", ttl)
}

func (l *limiter) IncreaseResp(ctx context.Context, remaining int, remoteIP net.IP, username string, reason string) (*passportAPI.AuthResponse, error) {
	b, err := broadcast.Driver()
	if err != nil {
		return nil, err
	}

	grpc.SendHeader(ctx, metadata.Pairs("Grpc-Metadata-X-RateLimit-Remaining", fmt.Sprintf("%d", remaining+1)))
	b.Publish("passport", broadcast.AuthenticateEvent{Event: &broadcast.Event{Type: "io.evntsrc.passport.limite_increased"}, Err: "limit increased", IP: remoteIP.String(), User: username})
	l.Inc(ctx, username, remoteIP)
	l.log.Printf("%s %s @ %s", reason, username, remoteIP)
	return &passportAPI.AuthResponse{Success: false}, status.Errorf(codes.Unauthenticated, "unknown username or password")
}
