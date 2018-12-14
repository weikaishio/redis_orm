package test

import (
	"testing"
	"time"

	"github.com/weikaishio/redis_orm"

	"github.com/go-redis/redis"
	"github.com/mkideal/log"
)

var (
	engine *redis_orm.Engine
)

//test and log?
func init() {
	options := redis.Options{
		Addr:               "127.0.0.1:6379",
		Password:           "",
		DB:                 1,
		DialTimeout:        10 * time.Second,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        60 * time.Second,
		IdleCheckFrequency: 15 * time.Second,
	}

	redisClient := redis.NewClient(&options)

	engine = redis_orm.NewEngine(redisClient)

	log.SetLevelFromString("TRACE")
}

//func TestEngine_GetTable(t *testing.T) {
//	faq := &Faq{
//		Title:   "为啥",
//		Content: "我也不知道",
//		Hearts:  20,
//	}
//	val := reflect.ValueOf(faq)
//	table, err := engine.GetTable(val)
//	t.Logf("table:%v,err:%v", table, err)
//}

func TestEngine_Insert(t *testing.T) {
	faq := &Faq{
		Hearts: 20,
	}
	err := engine.Insert(faq)
	t.Logf("Insert faq:%v,err:%v", faq, err)

	faq.Title = "titlexx"
	err = engine.Update(faq, "title")
	t.Logf("Update faq:%v, err:%v", faq, err)

	has, err := engine.Get(faq)
	t.Logf("Get faq:%v,has:%v,err:%v", faq, has, err)
}

//func TestEngine_Get(t *testing.T) {
//	faq := &Faq{
//		Id: 6,
//	}
//	has, err := engine.Get(faq)
//	t.Logf("faq:%v,has:%v,err:%v", faq, has, err)
//}

//func TestEngine_Update(t *testing.T) {
//	faq := &Faq{
//		Id:    5,
//		Title: "test5",
//	}
//	err := engine.Update(faq, "title")
//	t.Logf("TestEngine_Update err:%v", err)
//}

type Faq struct {
	Id        int64  `rds:"pk autoincr comment 'ID'"`
	Title     string `rds:"dft 'faqtitle' index comment '标题'"`
	Content   string `rds:"dft 'cnt' comment '内容'"`
	Hearts    int    `rds:"dft 10 comment '点赞数'"`
	CreatedAt int64  `rds:"created_at comment '创建时间'"`
	UpdatedAt int64  `rds:"updated_at comment '修改时间'"`
}
