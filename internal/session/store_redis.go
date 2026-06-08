package session

import (
	"context"
	"time"
)

// RedisStore Redis 会话存储（可选扩展）
// 需要时 go get github.com/redis/go-redis/v9 并取消注释
type RedisStore struct {
	addr     string
	password string
	db       int
	// client *redis.Client
	ctx context.Context
}

func NewRedisStore(addr, password string, db int) *RedisStore {
	s := &RedisStore{addr: addr, password: password, db: db, ctx: context.Background()}
	// TODO: 初始化 redis client
	// s.client = redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return s
}

func (s *RedisStore) Get(key string) (string, error) {
	// return s.client.Get(s.ctx, key).Result()
	return "", nil
}

func (s *RedisStore) Set(key, value string, ttl time.Duration) error {
	// return s.client.Set(s.ctx, key, value, ttl).Err()
	return nil
}

func (s *RedisStore) Delete(key string) error {
	// return s.client.Del(s.ctx, key).Err()
	return nil
}
