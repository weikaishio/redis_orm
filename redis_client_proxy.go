package redis_orm

import (
	"github.com/go-redis/redis"
)

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
	if len(fields) == 0 {
		return &redis.StatusCmd{}
	}
	val := c.redisClient.HMSet(key, fields)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) HIncrby(key string, field string, intVal int64) *redis.IntCmd {
	val := c.redisClient.HIncrBy(key, field, intVal)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) HIncrbyFloat(key string, field string, intVal float64) *redis.FloatCmd {
	val := c.redisClient.HIncrByFloat(key, field, intVal)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) HMGet(key string, fields ...string) *redis.SliceCmd {
	if len(fields) == 0 {
		return &redis.SliceCmd{}
	}
	val := c.redisClient.HMGet(key, fields...)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) HIncrBy(key, field string, incr int64) *redis.IntCmd {
	val := c.redisClient.HIncrBy(key, field, incr)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) HDel(key string, fields ...string) *redis.IntCmd {
	if len(fields) == 0 {
		return &redis.IntCmd{}
	}
	val := c.redisClient.HDel(key, fields...)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) Del(keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return &redis.IntCmd{}
	}
	val := c.redisClient.Del(keys...)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZCount(key, min, max string) *redis.IntCmd {
	val := c.redisClient.ZCount(key, min, max)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZScore(key, member string) *redis.FloatCmd {
	val := c.redisClient.ZScore(key, member)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	val := c.redisClient.ZRangeByScore(key, opt)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZRevRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	val := c.redisClient.ZRevRangeByScore(key, opt)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZRem(key string, members ...interface{}) *redis.IntCmd {
	if len(members) == 0 {
		return &redis.IntCmd{}
	}
	val := c.redisClient.ZRem(key, members...)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZRemRangeByScores(key string, min, max string) *redis.IntCmd {
	val := c.redisClient.ZRemRangeByScore(key, min, max)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	if len(members) == 0 {
		return &redis.IntCmd{}
	}
	val := c.redisClient.ZAdd(key, members...)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
func (c *RedisClientProxy) ZAddNX(key string, members ...redis.Z) *redis.IntCmd {
	if len(members) == 0 {
		return &redis.IntCmd{}
	}
	val := c.redisClient.ZAddNX(key, members...)
	c.engine.Printfln("========redis command========\n%v", val.String())
	return val
}
