package redisplus

import (
	"errors"
	"fmt"
	"gopkg.in/redis.v5"
)

func (r *redisView) ZLen(key string) (int64, error) {
	return r.cmd.ZCard(r.expandKey(key)).Result()
}

func (r *redisView) ZCount(key string, min, max float64) (int64, error) {
	return r.cmd.ZCount(r.expandKey(key),
		fmt.Sprintf("%f", min), fmt.Sprintf("%f", max)).Result()
}

func (r *redisView) ZLexCount(key, min, max string) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *redisView) ZAdd(key string, members ...*ZMember) (int64, error) {
	var zS []redis.Z
	for _, member := range members {
		z := redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
		zS = append(zS, z)
	}
	return r.cmd.ZAdd(r.expandKey(key), zS...).Result()
}

func (r *redisView) ZRem(key string, members ...*ZMember) (int64, error) {
	var zS []interface{}
	for _, member := range members {
		zS = append(zS, member.Member)
	}
	return r.cmd.ZRem(r.expandKey(key), zS...).Result()
}

func (r *redisView) ZRemRangeByLex(key, min, max string) (int64, error) {
	return r.cmd.ZRemRangeByLex(r.expandKey(key), min, max).Result()
}

func (r *redisView) ZRemRangeByScore(key string, min, max float64) (int64, error) {
	return r.cmd.ZRemRangeByScore(r.expandKey(key),
		fmt.Sprintf("%f", min), fmt.Sprintf("%f", max)).Result()
}

func (r *redisView) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return r.cmd.ZRemRangeByRank(r.expandKey(key), start, stop).Result()
}

func (r *redisView) ZRange(key string, start, stop int64, reverse, withScores bool) ([]*ZMember, error) {
	var err error
	var members []string
	var zSlice []redis.Z
	if !reverse && !withScores {
		members, err = r.cmd.ZRange(r.expandKey(key), start, stop).Result()
	}
	if reverse && withScores {
		zSlice, err = r.cmd.ZRevRangeWithScores(r.expandKey(key), start, stop).Result()
	}
	if reverse {
		members, err = r.cmd.ZRevRange(r.expandKey(key), start, stop).Result()
	}
	if withScores {
		zSlice, err = r.cmd.ZRangeWithScores(r.expandKey(key), start, stop).Result()
	}
	return toRangeZMembers(err, members, zSlice)
}

func (r *redisView) ZRangeByScore(key string, rangeBy ZRangeBy, reverse, withScores bool) ([]*ZMember, error) {
	var err error
	var members []string
	var zSlice []redis.Z
	if !reverse && !withScores {
		members, err = r.cmd.ZRangeByScore(r.expandKey(key), rangeBy.ToRedisRangeBy()).Result()
	}
	if reverse && withScores {
		zSlice, err = r.cmd.ZRevRangeByScoreWithScores(r.expandKey(key), rangeBy.ToRedisRangeBy()).Result()
	}
	if reverse {
		members, err = r.cmd.ZRevRangeByScore(r.expandKey(key), rangeBy.ToRedisRangeBy()).Result()
	}
	if withScores {
		zSlice, err = r.cmd.ZRangeByScoreWithScores(r.expandKey(key), rangeBy.ToRedisRangeBy()).Result()
	}
	return toRangeZMembers(err, members, zSlice)
}

func (r *redisView) ZRangeByLex(key string, rangeBy ZRangeBy, reverse bool) ([]*ZMember, error) {
	var err error
	var members []string
	if reverse {
		members, err = r.cmd.ZRevRangeByLex(key, rangeBy.ToRedisRangeBy()).Result()
	}
	members, err = r.cmd.ZRangeByLex(key, rangeBy.ToRedisRangeBy()).Result()
	return toRangeZMembers(err, members, []redis.Z{})
}

func (r *redisView) ZRank(key string, member []byte, reverse bool) (int64, error) {
	if reverse {
		return r.cmd.ZRevRank(r.expandKey(key), string(member)).Result()
	}
	return r.cmd.ZRank(r.expandKey(key), string(member)).Result()
}

func (r *redisView) ZIncr(key string, member *ZMember) (float64, error) {
	return r.cmd.ZIncr(r.expandKey(key), redis.Z{Score: member.Score, Member: member.Member}).Result()
}

func (r *redisView) ZIncrNX(key string, member *ZMember) (float64, error) {
	return r.cmd.ZIncrNX(r.expandKey(key), redis.Z{Score: member.Score, Member: member.Member}).Result()
}

func (r *redisView) ZInterMerge(destination string, merge *ZMerge, keys ...string) (int64, error) {
	var inKeys []string
	for _, key := range keys {
		inKeys = append(inKeys, r.expandKey(key))
	}
	return r.cmd.ZInterStore(destination, merge.ToZStore(), inKeys...).Result()
}

func (r *redisView) ZUnionMerge(destination string, merge *ZMerge, keys ...string) (int64, error) {
	var inKeys []string
	for _, key := range keys {
		inKeys = append(inKeys, r.expandKey(key))
	}
	return r.cmd.ZUnionStore(destination, merge.ToZStore(), inKeys...).Result()
}
