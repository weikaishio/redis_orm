package benchmark_test

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/weikaishio/redis_orm"
	"github.com/weikaishio/redis_orm/example/models"
	"testing"
	)

var (
	engine *redis_orm.Engine
)

//test and log?
func init() {
	options := redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	}

	redisClient := redis.NewClient(&options)

	engine = redis_orm.NewEngine(redisClient)

}

// go test -v -bench=".*"
/*
goos: darwin
goarch: amd64
pkg: github.com/weikaishio/redis_orm/test/benchmark_test
Benchmark_MysqlOrmGet-8            10000            197274 ns/op
Benchmark_MysqlOrmInsert-8          5000            259411 ns/op
Benchmark_RedisOrmGet-8            30000             44099 ns/op
Benchmark_RedisOrmInsert-8         10000            147920 ns/op
PASS
ok      github.com/weikaishio/redis_orm/test/benchmark_test     6.643s
*/
func Benchmark_RedisOrmGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		faq := &models.Faq{
			Id: int64(i),
		}
		engine.Get(faq)
	}
}

func Benchmark_RedisOrmInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		faq := &models.Faq{
			Title:   fmt.Sprintf("title%d", i),
			Content: fmt.Sprintf("contente%d", i),
			Hearts:  i,
		}
		engine.Insert(faq)
	}
}
