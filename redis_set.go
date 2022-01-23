package redisplus

func (r *redisView) SLen(key string) (int64, error) {
	return r.cmd.SCard(r.expandKey(key)).Result()
}

func (r *redisView) SAdd(key string, values ...[]byte) (int64, error) {
	var in []interface{}
	for _, value := range values {
		in = append(in, value)
	}
	return r.cmd.SAdd(r.expandKey(key), in...).Result()
}

func (r *redisView) SRem(key string, values ...[]byte) (int64, error) {
	var in []interface{}
	for _, value := range values {
		in = append(in, value)
	}
	return r.cmd.SRem(r.expandKey(key), in...).Result()
}

func (r *redisView) SPop(key string) ([]byte, error) {
	result, err := r.cmd.SPop(r.expandKey(key)).Result()
	if nil != err {
		return nil, err
	}
	return []byte(result), nil
}

func (r *redisView) SPopN(key string, count int64) ([][]byte, error) {
	return wrapSliceStringToSliceBytes(func() ([]string, error) {
		return r.cmd.SPopN(r.expandKey(key), count).Result()
	})
}

func (r *redisView) SDiff(keys ...string) ([][]byte, error) {
	var inkeys []string
	for _, key := range keys {
		inkeys = append(inkeys, r.expandKey(key))
	}

	return wrapSliceStringToSliceBytes(func() ([]string, error) {
		return r.cmd.SDiff(inkeys...).Result()
	})
}

func (r *redisView) SDiffMerge(destination string, keys ...string) (int64, error) {
	var inkeys []string
	for _, key := range keys {
		inkeys = append(inkeys, r.expandKey(key))
	}
	return r.cmd.SDiffStore(destination, inkeys...).Result()
}

func (r *redisView) SInter(keys ...string) ([][]byte, error) {
	var inkeys []string
	for _, key := range keys {
		inkeys = append(inkeys, r.expandKey(key))
	}
	return wrapSliceStringToSliceBytes(func() ([]string, error) {
		return r.cmd.SInter(inkeys...).Result()
	})
}

func (r *redisView) SInterMerge(destination string, keys ...string) (int64, error) {
	var inkeys []string
	for _, key := range keys {
		inkeys = append(inkeys, r.expandKey(key))
	}
	return r.cmd.SInterStore(destination, keys...).Result()
}

func (r *redisView) SUnion(keys ...string) ([][]byte, error) {
	var inkeys []string
	for _, key := range keys {
		inkeys = append(inkeys, r.expandKey(key))
	}
	return wrapSliceStringToSliceBytes(func() ([]string, error) {
		return r.cmd.SUnion(inkeys...).Result()
	})
}

func (r *redisView) SUnionMerge(destination string, keys ...string) (int64, error) {
	var inkeys []string
	for _, key := range keys {
		inkeys = append(inkeys, r.expandKey(key))
	}
	return r.cmd.SUnionStore(destination, inkeys...).Result()
}
