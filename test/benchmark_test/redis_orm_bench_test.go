package benchmark_test

import (
	"github.com/go-redis/redis"
	"github.com/mkideal/log"
	"github.com/weikaishio/redis_orm"
	"testing"
	"fmt"
	"github.com/weikaishio/redis_orm/test/models"
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

	log.SetLevelFromString("TRACE")
}

// go test -v -bench=".*"
func Benchmark_RedisOrmGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		faq := &models.Faq{
			Id: int64(i),
		}
		engine.Get(faq)
	}
}

func Benchmark_RedisOrmInsert(b *testing.B){
	for i := 0; i < b.N; i++ {
		faq:=&models.Faq{
			Title:fmt.Sprintf("title%d",i),
			Content:fmt.Sprintf("contente%d",i),
		}
		engine.Insert(faq)
	}
}