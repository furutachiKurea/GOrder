package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func SetNX(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration) (err error) {
	now := time.Now()
	defer func() {
		l := zerolog.Ctx(ctx).With().
			Time("start", now).
			Str("key", key).
			Str("value", value).
			Err(err).
			Int64("cost(ns)", time.Since(now).Nanoseconds()).Logger()

		if err == nil {
			l.Info().Msg("redis_setnx_success")
		} else {
			l.Warn().Msg("redis_setnx_error")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}

	_, err = client.SetNX(ctx, key, value, ttl).Result()
	return err
}

func Del(ctx context.Context, client *redis.Client, key string) (err error) {
	now := time.Now()
	defer func() {
		l := zerolog.Ctx(ctx).With().
			Time("start", now).
			Str("key", key).
			Err(err).
			Int64("cost(ns)", time.Since(now).Nanoseconds()).
			Logger()
		if err == nil {
			l.Info().Msg("redis_del_success")
		} else {
			l.Warn().Msg("redis_del_error")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}

	_, err = client.Del(ctx, key).Result()
	return err
}
