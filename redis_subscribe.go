package redisplus

import (
	"errors"
	"gopkg.in/redis.v5"
)

//Subscribe 订阅
//@return *redis.PubSub, error
func (r *redisView) Subscribe(channels ...string) (*redis.PubSub, error) {
	switch v := r.cmd.(type) {
	case *redis.Client:
		return v.Subscribe(channels...)
	default:
		return nil, errors.New("UnSupported")
	}
}

//PSubscribe  订阅
//channels ...string
//@return *redis.PubSub, error
func (r *redisView) PSubscribe(channels ...string) (*redis.PubSub, error) {
	switch v := r.cmd.(type) {
	case *redis.Client:
		return v.PSubscribe(channels...)
	default:
		return nil, errors.New("UnSupported")
	}
}
