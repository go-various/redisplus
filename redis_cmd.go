package redisplus

import (
	"errors"
	"gopkg.in/redis.v5"
)

var ErrRedisAddrsEmpty = errors.New("addrs is empty")


type RedisCmd interface {
	redis.Cmdable
}

func NewRedisCmd(c *Config) (RedisCmd, error) {
	if len(c.Addrs) == 0 {
		return nil, ErrRedisAddrsEmpty
	}

	if c.UseCluster {
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:              c.Addrs,
			ReadOnly:           c.ReadOnly,
			Password:           c.Password,
			DialTimeout:        c.DialTimeout,
			ReadTimeout:        c.ReadTimeout,
			WriteTimeout:       c.WriteTimeout,
			PoolSize:           c.PoolSize,
			PoolTimeout:        c.PoolTimeout,
			IdleTimeout:        c.IdleTimeout,
			IdleCheckFrequency: c.IdleCheckFrequency,
		}), nil
	}
	return redis.NewClient(&redis.Options{
		Addr:               c.Addrs[0],
		Password:           c.Password,
		DB:                 c.DbIndex,
		DialTimeout:        c.DialTimeout,
		ReadTimeout:        c.ReadTimeout,
		WriteTimeout:       c.WriteTimeout,
		PoolSize:           c.PoolSize,
		PoolTimeout:        c.PoolTimeout,
		IdleTimeout:        c.IdleTimeout,
		IdleCheckFrequency: c.IdleCheckFrequency,
		ReadOnly:           c.ReadOnly,
		TLSConfig:          c.TLSConfig,
	}), nil
}
