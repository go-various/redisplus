// redis expired notification

package redisplus

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var errPrefixNotNil = errors.New("prefix must be not null")
var errRedisNotNil = errors.New("redis must be not null")

const NotifyKeyPrefix = "NOTIFY"
const NotifyValuePrefix = "NOTIFY_VALUE"
const NotifyLockPrefix = "NOTIFY_LOCK"

type PutNext bool
type NotificationHandler func(p *Entity, err error) PutNext

func DefaultRetryPolicies() []time.Duration {
	return []time.Duration{
		time.Minute * 1,
		time.Minute * 5,
		time.Minute * 10,
		time.Minute * 30,
		time.Minute * 60,
		time.Minute * 120,
	}
}

type Entity struct {
	count int64
	Key   string
	Value []byte
}

//通知key
func (e *Entity) notifyKey() string {
	key := base64.StdEncoding.EncodeToString([]byte(e.Key))
	return fmt.Sprintf("%s:%s||%d", NotifyKeyPrefix, key, e.count)
}

//通知 value key
func (e *Entity) valueKey() string {
	key := base64.StdEncoding.EncodeToString([]byte(e.Key))
	return fmt.Sprintf("%s:%s", NotifyValuePrefix, key)
}

type Notification interface {
	PutNotification(p *Entity) error
	Subscribe(handler NotificationHandler) error
}

type policies []time.Duration

func (receiver policies) last() time.Duration {
	if len(receiver) == 0{
		return 0
	}
	return receiver[len(receiver)-1]
}
func (receiver policies) length() int {
	return len(receiver)
}

type notification struct {
	prefix   string
	node     string
	cache    RedisCli
	policies policies
	logger  *log.Logger
}

func NewNotification(prefix string, cache RedisCli, logger *log.Logger,policies ...[]time.Duration) (Notification, error) {

	if "" == prefix {
		return nil, errPrefixNotNil
	}
	if nil == cache {
		return nil, errRedisNotNil
	}

	n := &notification{
		prefix: prefix,
		node:   fmt.Sprintf("%s:%d", GetLocalIP(), os.Getpid()),
		cache:  cache,
		logger: logger,
	}

	if 0 == len(policies) {
		n.policies = DefaultRetryPolicies()
	} else {
		n.policies = policies[0]
	}

	return n, nil
}

func (n *notification) PutNotification(p *Entity) error {
	ctx := context.Background()
	if err := n.cache.Set(ctx, n.prefix+":"+p.notifyKey(), p.Value, n.policies.last().String()); err != nil {
		return err
	}
	//设置value的最大缓存过期时间为重试policy的最大时间
	if err := n.cache.Set(ctx, n.prefix+":"+p.valueKey(), p.Value, n.policies.last().String()); err != nil {
		return err
	}
	return nil
}

func (n *notification) Subscribe(handler NotificationHandler) error {
	space := fmt.Sprintf("__keyspace@*__:%s:%s*", n.cache.KeyPrefix(), n.prefix)

	psub, err := n.cache.PSubscribe(space)
	if nil != err {
		return err
	}

	go func() {
		for {
			message, err := psub.ReceiveMessage()
			if nil != err {
				continue
			}

			if message.Payload == "expired" {
				entity, err := n.decode(message.Channel)
				if nil == entity {
					continue
				}
				if !n.lock(entity) {
					continue
				}
				entity.count += 1
				err = n.fetchValue(entity)
				putNext := handler(entity, err)
				if putNext && entity.count < int64(n.policies.length()) {
					if err := n.PutNotification(entity); err != nil {
						n.logger.Println("put notification", "err", err)
						continue
					}
				}
				if !putNext || (putNext && entity.count >= int64(n.policies.length())) {
					n.unlock(entity)
				}
			}
		}
	}()
	return nil
}

func (n *notification) decode(src string) (*Entity, error) {
	//N:N:N:B64||count
	keys := strings.Split(src, ":")
	if len(keys) < 1 {
		return nil, errors.New("key format error: " + src)
	}
	//B64||count
	keyCount := strings.Split(keys[len(keys)-1], "||")
	if len(keyCount) != 2 {
		return nil, errors.New("key format error: " + src)
	}
	keyBytes, err := base64.StdEncoding.DecodeString(keyCount[0])
	if nil != err {
		return nil, errors.New("key format error: " + src)
	}
	count, _ := strconv.ParseInt(keyCount[1], 10, 32)
	en := &Entity{
		count: count,
		Key:   string(keyBytes),
	}
	return en, nil
}

func (n *notification) fetchValue(entity *Entity) error {
	ctx := context.Background()
	value, err := n.cache.Get(ctx, entity.valueKey())
	entity.Value = value
	return err
}

//锁定通知防止多实例处理冲突
func (n *notification) lock(p *Entity) bool {
	//"${NOTIFY_PREFIX}:NOTIFY_LOCK:15ba4ad6-5923-4a9d-89c9-b35f33c60fa3"
	setKey := fmt.Sprintf("%s:%s:%s",n.prefix, NotifyLockPrefix,  p.Key)
	value := fmt.Sprintf("%s:%s:%d", NotifyLockPrefix, n.node, GetRoutineID())
	ret, _ := n.cache.SetNX(context.Background(), setKey, []byte(value), n.policies.last().String())
	if ret {
		return ret
	}
	rVal, _ := n.cache.Get(context.Background(), setKey)
	return value == string(rVal)
}

//解锁通知实例
func (n *notification) unlock(p *Entity) {
	setKey := fmt.Sprintf("%s:%s:%s", n.prefix, NotifyLockPrefix,  p.Key)
	n.cache.Del(context.Background(), setKey)
}
