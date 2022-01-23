package redisplus

import (
	"crypto/tls"
	"gopkg.in/redis.v5"
	"time"
)

const RedisKeySep = ":"

type Config struct {
	Addrs        []string      `hcl:"addrs"`      //集群地址
	Password     string        `hcl:"password"`   //密码
	KeyPrefix    string        `hcl:"key_prefix"` //key前缀
	DbIndex      int           `hcl:"db_index"`   //数据索引
	DialTimeout  time.Duration `hcl:"dial_timeout"`
	ReadTimeout  time.Duration `hcl:"read_timeout"`
	WriteTimeout time.Duration `hcl:"write_timeout"`
	ReadOnly     bool          `hcl:"read_only"`
	// PoolSize applies per cluster node and not for the whole cluster.
	PoolSize           int           `hcl:"pool_size"`
	PoolTimeout        time.Duration `hcl:"pool_timeout"`
	IdleTimeout        time.Duration `hcl:"idle_timeout"`
	IdleCheckFrequency time.Duration `hcl:"idle_check_frequency"`
	UseCluster         bool          `hcl:"use_cluster"`
	TLSConfig          *tls.Config   `hcl:"tls_config"`
}

type RedisCli interface {
	KeyPrefix() string

	Scan(cursor uint64, match string, count int64) ([]string, error)

	// SetNX Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	SetNX(key string, value []byte, duration string) (bool, error)
	Get(key string) ([]byte, error)
	// Set Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Set(key string, value []byte, duration string) error
	Del(keys ...string) (int64, error)
	Expire(key string, duration string) error
	HSetNX(key, field string, value []byte) error
	HSet(key, field string, value []byte) error
	HMSet(key string, Values map[string][]byte) error
	HGet(key, field string) ([]byte, error)
	HMGet(key string, fields ...string) ([][]byte, error)
	HGetAll(key string) (map[string][]byte, error)
	HDel(key string, fields ...string) (int64, error)
	HLen(key string) (int64, error)
	HKeys(key string) ([]string, error)
	HValues(key string) ([][]byte, error)
	HExists(key, field string) (bool, error)
	LRem(key string, count int64, value []byte) (int64, error)
	LIndex(key string, index int64) ([]byte, error)
	LTrim(key string, start, stop int64) error
	LSet(key string, index int64, value []byte) error
	LPush(key string, values ...[]byte) (int64, error)
	LAppend(key string, values ...[]byte) (int64, error)
	LPop(key string) ([]byte, error)
	LRPop(key string) ([]byte, error)
	LRange(key string, start, stop int64) ([][]byte, error)
	LLen(key string) (int64, error)
	LInsert(key string, op InsertOP, pivot, value []byte) (int64, error)
	SLen(key string) (int64, error)
	SAdd(key string, values ...[]byte) (int64, error)
	SRem(key string, values ...[]byte) (int64, error)
	SPop(key string) ([]byte, error)
	SPopN(key string, count int64) ([][]byte, error)
	SDiff(keys ...string) ([][]byte, error)
	SDiffMerge(destination string, keys ...string) (int64, error)
	SInter(keys ...string) ([][]byte, error)
	SInterMerge(destination string, keys ...string) (int64, error)
	SUnion(keys ...string) ([][]byte, error)
	SUnionMerge(destination string, keys ...string) (int64, error)
	ZLen(key string) (int64, error)
	ZCount(key string, min, max float64) (int64, error)
	ZLexCount(key, min, max string) (int64, error)
	ZAdd(key string, members ...*ZMember) (int64, error)
	ZRem(key string, members ...*ZMember) (int64, error)
	ZRemRangeByLex(key, min, max string) (int64, error)
	ZRemRangeByScore(key string, min, max float64) (int64, error)
	ZRemRangeByRank(key string, start, stop int64) (int64, error)

	ZRange(key string, start, stop int64, reverse, withScores bool) ([]*ZMember, error)

	ZRangeByScore(key string, rangeBy ZRangeBy, reverse, withScores bool) ([]*ZMember, error)

	ZRangeByLex(key string, rangeBy ZRangeBy, reverse bool) ([]*ZMember, error)

	ZRank(key string, member []byte, reverse bool) (int64, error)

	ZIncr(key string, member *ZMember) (float64, error)
	ZIncrNX(key string, member *ZMember) (float64, error)

	ZInterMerge(destination string, merge *ZMerge, keys ...string) (int64, error)
	ZUnionMerge(destination string, merge *ZMerge, keys ...string) (int64, error)

	GeoAdd(key string, geoLocation ...*redis.GeoLocation) (int64, error)
	GeoRadius(key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error)
	GeoRadiusByMember(key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error)
	GeoDist(key string, member1, member2, unit string) (float64, error)
	GeoHash(key string, members ...string) ([]string, error)
	GeoPos(key string, members ...string) ([]*redis.GeoPos, error)
	GeoCalculateDistance(key string, location1 Location, location2 Location) (float64, error)

	Subscribe(channels ...string) (*redis.PubSub, error)
	PSubscribe(channels ...string) (*redis.PubSub, error)

	NativeCmd() RedisCmd
}

type InsertOP int32

const (
	InsertOP_BEFORE InsertOP = 0
	InsertOP_AFTER  InsertOP = 1
)

type RedisEntry struct {
	Key   string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

// ZMember map
type ZMember struct {
	Score  float64 `protobuf:"fixed32,1,opt,name=score,proto3" json:"score,omitempty"`
	Member []byte  `protobuf:"bytes,2,opt,name=member,proto3" json:"member,omitempty"`
}

type ZRangeBy struct {
	Min    string `protobuf:"bytes,1,opt,name=min,proto3" json:"min,omitempty"`
	Max    string `protobuf:"bytes,2,opt,name=max,proto3" json:"max,omitempty"`
	Offset int64  `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	Count  int64  `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
}

func (z ZRangeBy) ToRedisRangeBy() redis.ZRangeBy {
	return redis.ZRangeBy{
		Min:    z.Min,
		Max:    z.Max,
		Offset: z.Offset,
		Count:  z.Count,
	}
}

// ZMerge used as an arg to ZInterMerge and ZUnionMerge.
type ZMerge struct {
	Weights   []float64 `protobuf:"fixed64,1,rep,packed,name=Weights,proto3" json:"Weights,omitempty"`
	Aggregate string    `protobuf:"bytes,2,opt,name=Aggregate,proto3" json:"Aggregate,omitempty"` //Can be SUM, MIN or MAX.
}

func (z ZMerge) ToZStore() redis.ZStore {
	return redis.ZStore{
		Weights:   z.Weights,
		Aggregate: z.Aggregate,
	}
}
