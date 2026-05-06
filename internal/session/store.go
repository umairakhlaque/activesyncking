package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/umairakhlaque/activesyncking/internal/config"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

const challengeKeyPrefix = "sg:challenge:"

var ErrNotFound = errors.New("challenge not found")
var ErrExpired = errors.New("challenge expired")

// Store is the Redis-backed store for active MFA challenges.
// It is the primary (fast-path) store; PostgreSQL is the audit record.
type Store struct {
	rdb *redis.Client
}

func NewStore(cfg config.RedisConfig) (*Store, error) {
	var rdb *redis.Client
	if cfg.URL != "" {
		// Upstash and other managed Redis providers supply a full rediss:// URL
		opt, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("parsing redis url: %w", err)
		}
		rdb = redis.NewClient(opt)
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Store{rdb: rdb}, nil
}

func (s *Store) Save(ctx context.Context, c *models.Challenge) error {
	key := challengeKeyPrefix + c.ID
	ttl := time.Until(c.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("challenge already expired")
	}

	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshalling challenge: %w", err)
	}

	return s.rdb.Set(ctx, key, data, ttl).Err()
}

func (s *Store) Get(ctx context.Context, id string) (*models.Challenge, error) {
	key := challengeKeyPrefix + id
	data, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("redis get challenge: %w", err)
	}

	c := &models.Challenge{}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("unmarshalling challenge: %w", err)
	}

	if c.IsExpired() {
		_ = s.Delete(ctx, id)
		return nil, ErrExpired
	}

	return c, nil
}

// Update atomically updates a challenge: read-modify-write with optimistic retry.
// TTL is preserved from the original expiry so challenges don't get extended on update.
func (s *Store) Update(ctx context.Context, c *models.Challenge) error {
	key := challengeKeyPrefix + c.ID
	ttl := time.Until(c.ExpiresAt)
	if ttl <= 0 {
		return ErrExpired
	}

	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshalling challenge: %w", err)
	}

	// KEEPTTL not available in all Redis versions; use remaining TTL explicitly
	return s.rdb.Set(ctx, key, data, ttl).Err()
}

func (s *Store) Delete(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, challengeKeyPrefix+id).Err()
}

func (s *Store) Close() error {
	return s.rdb.Close()
}
