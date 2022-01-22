package redisplus

import (
	"context"
	"errors"
	"gopkg.in/redis.v5"
	"time"
)

var ErrorResultNotOK = errors.New("result is not OK")
var ErrorResultNotTrue = errors.New("result is not true")
var ErrorInputValuesIsNil = errors.New("input values is nil")

const RedisKeySep = ":"

var _ RedisCli = (*redisView)(nil)

type redisView struct {
	prefix string
	cmd    RedisCmd
	pubSub redis.PubSub
}

func NewRedisView(cmd RedisCmd, prefix string) RedisCli {
	view := &redisView{cmd: cmd, prefix: prefix}
	return view
}

func (r *redisView) KeyPrefix() string {
	return r.prefix
}
func (r *redisView) NativeCmd() RedisCmd {
	return r.cmd
}

func (r *redisView) SetNX(ctx context.Context, key string, value []byte, duration string) (bool, error) {
	if duration != "" {
		timeout, err := time.ParseDuration(duration)
		if nil != err {
			return false, err
		}
		return r.cmd.SetNX(r.expandKey(key), value, timeout).Result()
	}
	return r.cmd.SetNX(r.expandKey(key), value, 0).Result()
}

func (r *redisView) Scan(cursor uint64, match string, count int64) ([]string, error) {
	result, _, err := r.cmd.Scan(cursor, r.expandKey(match), count).Result()
	if nil != err {
		return nil, err
	}
	return result, err
}

func (r *redisView) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.cmd.Get(r.expandKey(key)).Result()
	if nil != err {
		return nil, err
	}
	return []byte(result), nil
}

func (r *redisView) Set(ctx context.Context, key string, value []byte, duration string) error {
	if duration != "" {
		timeout, err := time.ParseDuration(duration)
		if nil != err {
			return err
		}
		return r.cmd.Set(r.expandKey(key), value, timeout).Err()
	}
	return r.cmd.Set(r.expandKey(key), value, 0).Err()
}

func (r *redisView) Del(ctx context.Context, keys ...string) (int64, error) {
	var all []string
	for _, key := range keys {
		all = append(all, r.expandKey(key))
	}
	return r.cmd.Del(all...).Result()
}

func (r *redisView) Expire(ctx context.Context, key string, duration string) error {
	timeout, err := time.ParseDuration(duration)
	if nil != err {
		return err
	}
	return wrapResult(func() (interface{}, error) {
		return r.cmd.Expire(r.expandKey(key), timeout).Result()
	})
}
