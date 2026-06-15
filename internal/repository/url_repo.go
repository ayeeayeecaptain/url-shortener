package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type URLRepository struct {
	db       *sql.DB
	rdb      *redis.Client
	cacheTTL time.Duration
}

func NewURLRepository(db *sql.DB, rdb *redis.Client) *URLRepository {
	return &URLRepository{
		db:       db,
		rdb:      rdb,
		cacheTTL: 24 * time.Hour, // Evict unclicked items from cache after a day
	}
}

// Save inserts the long URL and assigns it a unique serial tracking ID
func (r *URLRepository) Save(ctx context.Context, longURL string) (uint64, error) {
	var id uint64
	query := `INSERT INTO urls (long_url) VALUES ($1) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, longURL).Scan(&id)
	return id, err
}

// GetLongURL resolves URLs via Cache-Aside strategy
func (r *URLRepository) GetLongURL(ctx context.Context, token string, id uint64) (string, error) {
	// 1. Evaluate Cache Layer
	longURL, err := r.rdb.Get(ctx, token).Result()
	if err == nil {
		return longURL, nil // Cache Hit
	}

	if !errors.Is(err, redis.Nil) {
		return "", err // Fail safe if structural Redis network errors occur
	}

	// 2. Cache Miss: Execute query against PostgreSQL
	query := `SELECT long_url FROM urls WHERE id = $1`
	err = r.db.QueryRowContext(ctx, query, id).Scan(&longURL)
	if err != nil {
		return "", err
	}

	// 3. Heal Cache: Write missing back asynchronously to prioritize user response time
	go func() {
		_ = r.rdb.Set(context.Background(), token, longURL, r.cacheTTL).Err()
	}()

	return longURL, nil
}
