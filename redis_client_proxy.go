package redis_orm

import "github.com/go-redis/redis"

type RedisClientProxy struct {
	engine      *Engine
	redisClient *redis.Client
}

func NewRedisCliProxy(redisCli *redis.Client) *RedisClientProxy {
	return &RedisClientProxy{
		redisClient: redisCli,
	}
}
func (c *RedisClientProxy) HMSet(key string, fields map[string]interface{}) *redis.StatusCmd {
	val := c.redisClient.HMSet(key, fields)
	c.engine.Printfln("hmset(%s,%v) val:%v", key, fields, *val)
	return val
}
func (c *RedisClientProxy) HMGet(key string, fields ...string) *redis.SliceCmd {
	val := c.redisClient.HMGet(key, fields...)
	c.engine.Printfln("hmget(%s,%v) val:%v", key, fields, *val)
	return val
}
func (c *RedisClientProxy) HIncrBy(key, field string, incr int64) *redis.IntCmd {
	val := c.redisClient.HIncrBy(key, field, incr)
	c.engine.Printfln("hincrby(%s,%v,%d) val:%v", key, field, incr, val)
	return val
}
func (c *RedisClientProxy) HDel(key string, fields ...string) *redis.IntCmd {
	val := c.redisClient.HDel(key, fields...)
	c.engine.Printfln("hdel(%s,%v) val:%v", key, fields, *val)
	return val
}
func (c *RedisClientProxy) Del(keys ...string) *redis.IntCmd {
	val := c.redisClient.Del(keys...)
	c.engine.Printfln("del(%v) val:%v", keys, *val)
	return val
}
func (c *RedisClientProxy) ZCount(key, min, max string) *redis.IntCmd {
	val := c.redisClient.ZCount(key, min, max)
	c.engine.Printfln("zcount(%s,%v,%v) val:%v", key, min, max, *val)
	return val
}
func (c *RedisClientProxy) ZScore(key, member string) *redis.FloatCmd {
	val := c.redisClient.ZScore(key, member)
	c.engine.Printfln("zscore(%s,%v) val:%v", key, member, *val)
	return val
}
func (c *RedisClientProxy) ZRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	val := c.redisClient.ZRangeByScore(key, opt)
	c.engine.Printfln("zrangebyscore(%s,%v) val:%v", key, opt, *val)
	return val
}
func (c *RedisClientProxy) ZRevRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	val := c.redisClient.ZRevRangeByScore(key, opt)
	c.engine.Printfln("zrevrangebyscore(%s,%v) val:%v", key, opt, *val)
	return val
}
func (c *RedisClientProxy) ZRem(key string, members ...interface{}) *redis.IntCmd {
	val := c.redisClient.ZRem(key, members...)
	c.engine.Printfln("zrem(%s,%v) val:%v", key, members, *val)
	return val
}
func (c *RedisClientProxy) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	val := c.redisClient.ZAdd(key, members...)
	c.engine.Printfln("zadd(%s,%v) val:%v", key, members, *val)
	return val
}
func (c *RedisClientProxy) ZAddNX(key string, members ...redis.Z) *redis.IntCmd {
	val := c.redisClient.ZAddNX(key, members...)
	c.engine.Printfln("zaddnx(%s,%v) val:%v", key, members, *val)
	return val
}
