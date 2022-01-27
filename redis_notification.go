// redis expired notification

package redisplus

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var errPrefixNotNil = errors.New("prefix must be not null")
var errRedisNotNil = errors.New("redis must be not null")
var errKeyFormat = errors.New("key format error")

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
	return fmt.Sprintf("%s:%s:%d", NotifyKeyPrefix, key, e.count)
}

//通知 value key
func (e *Entity) valueKey() string {
	key := base64.StdEncoding.EncodeToString([]byte(e.Key))
	return fmt.Sprintf("%s:%s", NotifyValuePrefix, key)
}

func (e *Entity) String()string{
	bs, _ := json.Marshal(e)
	return string(bs)
}

type Notification interface {
	PutNotification(p *Entity) error
	Subscribe(handler NotificationHandler) error
}

type policies []time.Duration

func (po policies) last() time.Duration {
	if len(po) == 0 {
		return 0
	}
	return po[len(po)-1]
}
func (po policies) length() int {
	return len(po)
}

func (po policies) holding() time.Duration  {
	if len(po) == 0 {
		return 0
	}
	return po[len(po)-1] + time.Duration(5) * time.Minute
}

type notification struct {
	prefix   string
	node     string
	cache    RedisCli
	policies policies
	logger   Logger
}

func NewNotification(prefix string, cache RedisCli, logger Logger, policies ...[]time.Duration) (Notification, error) {

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
		err := checkPoliciesSequence(n)
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

//checkPoliciesSequence 检查策略是否为一个升序序列
func checkPoliciesSequence(n *notification) error{
	if len(n.policies) > 1 {
		for idx, policy := range n.policies[:len(n.policies)-1] {
			if policy > n.policies[idx+1] {
				return errors.New("policies must be a sequence of Duration")
			}
		}
	}
	return nil
}

func (n *notification) PutNotification(p *Entity) error {
	if err := n.cache.Set(n.prefix+":"+p.notifyKey(), []byte{}, n.policies[p.count].String()); err != nil {
		return err
	}

	//设置value的最大缓存过期时间为重试policy的最大时间
	if err := n.cache.Set(n.prefix+":"+p.valueKey(), p.Value, n.policies.holding().String()); err != nil {
		return err
	}
	return nil
}

func (n *notification) Subscribe(handler NotificationHandler) error {
	space := fmt.Sprintf("__keyspace@*__:%s:%s:%s:*", n.cache.KeyPrefix(), n.prefix, NotifyKeyPrefix)
	n.logger.Info("psubscribe key", "pattern", space)
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
			if message.Payload != "expired" {
				continue
			}

			entity, err := n.decode(message.Channel)
			if nil != err {
				n.logger.Error("decode message", "key",message.Channel, "err", err)
				continue
			}

			locked, err :=n.lock(entity)
			if nil != err {
				n.logger.Error("lock entity", "key", entity, "err", err)
				continue
			}
			//locked by another process
			if !locked{
				continue
			}

			err = n.fetchValue(entity)
			putNext := handler(entity, err)
			entity.count += 1
			if putNext && entity.count < int64(n.policies.length()) {
				if err := n.PutNotification(entity); err != nil {
					n.logger.Error("requeue entity", "entity", entity, "err", err)
					continue
				}
			}
			if !putNext || (putNext && entity.count >= int64(n.policies.length())) {
				n.unlock(entity)
			}
		}
	}()
	return nil
}

//decode value from TEST:dev:order:NOTIFY:NjI0YjVhNWYtNjA5Ny00YTgzLTkxMWYtNmU2N2NhYjZlOWJh:1
//return Entity
func (n *notification) decode(src string) (*Entity, error) {
	keys := strings.Split(src, ":")
	length := len(keys)
	if length < 4 {
		return nil,errKeyFormat
	}
	keyBytes, err := base64.StdEncoding.DecodeString(keys[length-2])
	if nil != err {
		return nil, errKeyFormat
	}
	count, err := strconv.ParseInt(keys[length-1], 10, 32)
	if err != nil {
		return nil, errKeyFormat
	}
	en := &Entity{
		count: count,
		Key:   string(keyBytes),
	}
	return en, nil
}

//fetchValue
//"TEST:dev:order:NOTIFY_VALUE:ZmY3ZDg5NGYtZjcxNi00NzlhLTk1YWYtOWViZjExYWUzZTEx"
func (n *notification) fetchValue(entity *Entity) error {
	key := fmt.Sprintf("%s:%s", n.prefix, entity.valueKey())
	value, err := n.cache.Get(key)
	entity.Value = value
	return err
}

//锁定通知防止多实例处理冲突
func (n *notification) lock(p *Entity) (bool, error) {
	//"${NOTIFY_PREFIX}:NOTIFY_LOCK:15ba4ad6-5923-4a9d-89c9-b35f33c60fa3"
	setKey := fmt.Sprintf("%s:%s:%s", n.prefix, NotifyLockPrefix, p.notifyKey())
	value := fmt.Sprintf("%s:%s", NotifyLockPrefix, n.node)
	ret, err := n.cache.SetNX(setKey, []byte(value), n.policies.last().String())
	if err != nil {
		return false, err
	}
	if !ret{
		return false, nil
	}
	rVal, err := n.cache.Get(setKey)
	if err != nil {
		return false, err
	}

	return  value == string(rVal), nil
}

//解锁通知实例
func (n *notification) unlock(p *Entity) {
	setKey := fmt.Sprintf("%s:%s:%s", n.prefix, NotifyLockPrefix, p.notifyKey())
	n.cache.Del(setKey)
}
