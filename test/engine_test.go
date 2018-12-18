package test

import (
	"github.com/weikaishio/redis_orm"
	"testing"

	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/mkideal/log"
	"github.com/weikaishio/redis_orm/test/models"
	"time"
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
	engine.IsShowLog(true)

	log.SetLevelFromString("TRACE")
}

//func TestEngine_GetTable(t *testing.T) {
//	faq := &models.Faq{
//		Title:   "为啥",
//		Content: "我也不知道",
//		Hearts:  20,
//	}
//	val := reflect.ValueOf(faq)
//	table, err := engine.GetTable(val)
//	t.Logf("table:%v,err:%v", table, err)
//}

func TestEngine_Insert(t *testing.T) {
	faq := &models.Faq{
		Title:  "index",
		Unique: time.Now().Unix(),
	}
	err := engine.Insert(faq)
	bys, _ := json.Marshal(faq)
	t.Logf("Insert faq:%v,err:%v", string(bys), err)

	//faq.Title = "titlexx"
	//err := engine.Update(faq, "title")
	//t.Logf("Update faq:%v, err:%v", faq, err)

	//has, err := engine.GetByCondition(faq, &redis_orm.SearchCondition{
	//	SearchColumn:  []string{"Type", "Hearts"},
	//	IndexType:     redis_orm.IndexType_IdMember,
	//	FieldMinValue: 1<<32 | 20,
	//	FieldMaxValue: 1<<32 | 20,
	//})
	//t.Logf("Get faq:%v,has:%v,err:%v", faq, has, err)
}

func TestEngine_Get(t *testing.T) {
	faq := &models.Faq{
		Id: 2,
	}
	has, err := engine.Get(faq)
	bys, _ := json.Marshal(faq)
	t.Logf("faq:%v,has:%v,err:%v", string(bys), has, err)
}

func TestEngine_GetByCombinedIndex(t *testing.T) {
	faq := &models.Faq{}
	has, err := engine.GetByCondition(faq, redis_orm.NewSearchCondition(
		redis_orm.IndexType_IdScore,
		"1&index",
		"1&index",
		"Type",
		"Title",
	))

	bys, _ := json.Marshal(faq)
	t.Logf("faq:%v,has:%v,err:%v", string(bys), has, err)
}

func TestEngine_Find(t *testing.T) {
	var faqAry []*models.Faq
	count, err := engine.Find(0, 3, redis_orm.NewSearchCondition(
		redis_orm.IndexType_IdScore,
		"-inf",
		"+inf",
		"Id",
	), &faqAry)
	bys, _ := json.Marshal(faqAry)
	t.Logf("faqAry:%v,count:%v,err:%v", string(bys), count, err)
}

func TestEngine_Update(t *testing.T) {
	faq := &models.Faq{
		Id:    5,
		Title: "test5",
	}
	err := engine.Update(faq, "title")
	t.Logf("TestEngine_Update err:%v", err)
}

func TestEngine_Delete(t *testing.T) {
	faq := &models.Faq{
		Id: 1,
	}
	err := engine.Delete(faq)
	t.Logf("Delete faq:%v, err:%v", faq, err)
}

func TestEngine_TableDrop(t *testing.T){
	faq:=&models.Faq{}
	err:=engine.TableDrop(faq)
	t.Logf("TestEngine_TableDrop err:%v", err)
}
