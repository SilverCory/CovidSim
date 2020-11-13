package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/SilverCory/CovidSim"
	"github.com/go-redis/redis/v8"
	"image"
	"time"
)

const (
	KeyHasMask = "covid:has_mask#"
	TTLHasMask = time.Hour

	KeySetMaskCooldown = "covid:mask_cooldown#"
	TTLSetMaskCooldown = time.Minute * 10
)

var _ CovidSim.Cache = &Redis{}

type Redis struct {
	client *redis.Client
}

func NewRedis(addr, username, password string, db int) (*Redis, error) {
	var ret = new(Redis)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	})

	res := rdb.Ping(context.Background())
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("NewRedis ping db: %w", err)
	}

	ret.client = rdb
	return ret, nil
}

func (r *Redis) InvalidateUser(userID string) error {
	if err := r.client.Del(context.TODO(), KeyHasMask+userID).Err(); err != nil {
		return fmt.Errorf("invalidate user: %w", err)
	}
	return nil
}

func (r *Redis) AvatarCache(userID string, getAvatarFunc func(userID string) (image.Image, error)) (image.Image, error) {
	return getAvatarFunc(userID)
}

func (r *Redis) HasMaskCache(userID string, hasMaskFunc func(userID string) (bool, error)) (bool, error) {
	var key = KeyHasMask + userID

	result := r.client.Get(context.TODO(), key)
	err := result.Err()

	if errors.Is(err, redis.Nil) {
		hasMask, err := hasMaskFunc(userID)
		if err != nil {
			return false, fmt.Errorf("HasMaskCache: unable to call hasMaskFunc for user %s: %w", userID, err)
		}

		err = r.client.Set(context.TODO(), key, hasMask, TTLHasMask).Err()
		if err != nil {
			return false, fmt.Errorf("HasMaskCache: unable to store result for user %s: %w", userID, err)
		}

		return hasMask, nil
	} else if err != nil {
		return false, fmt.Errorf("HasMaskCache: unable to get from redis cache: %w", err)
	}

	var hasMask bool
	if err := result.Scan(&hasMask); err != nil {
		return false, fmt.Errorf("HasMaskCache: scan result failed: %w", err)
	}

	return hasMask, nil
}

func (r *Redis) SetMaskCooldown(userID string) error {
	var key = KeySetMaskCooldown + userID
	err := r.client.Set(context.TODO(), key, true, TTLSetMaskCooldown).Err()
	if err != nil {
		return fmt.Errorf("HasMaskCache: unable to store cooldown for user %s: %w", userID, err)
	}
	return nil
}

func (r *Redis) GetMaskCooldown(userID string) (bool, error) {
	var key = KeySetMaskCooldown + userID

	result := r.client.Get(context.TODO(), key)
	err := result.Err()

	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("HasMaskCache: unable to get from redis cache: %w", err)
	}

	return true, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
