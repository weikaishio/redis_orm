package test

import (
		"github.com/go-redis/redis"
	"github.com/weikaishio/redis_orm"
	"github.com/weikaishio/redis_orm/example/models"
	"testing"
)

var (
	engine *redis_orm.Engine
)

//test and log? use printf
func init() {
	options := redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	}

	redisClient := redis.NewClient(&options)

	engine = redis_orm.NewEngine(redisClient)
	engine.IsShowLog(true)
	_, err := engine.Schema.ReloadTables()
	if err != nil {
		engine.Printfln("ReloadTables err:%v", err)
	}
}

//func TestEngine_CreateTable(t *testing.T) {
//	faq := &models.Faq{}
//	engine.Schema.TableDrop(faq)
//	err := engine.Schema.CreateTable(faq)
//	t.Logf("CreateTable(%v),err:%v", faq, err)
//}

//func TestEngine_GetTable(t *testing.T) {
//	faq := &models.Faq{
//		Title:   "为啥",
//		Content: "我也不知道",
//		Hearts:  20,
//	}
//	val := reflect.ValueOf(faq)
//	table, err := engine.GetTable(val, reflect.Indirect(val))
//	t.Logf("table:%v,err:%v", table, err)
//}

//func TestEngine_GetDefaultValue(t *testing.T) {
//	faq := &models.Faq{}
//	err := engine.GetDefaultValue(faq)
//	bys, _ := json.Marshal(faq)
//	t.Logf("GetDefaultValue faq:%v,err:%v", string(bys), err)
//}
//
func TestEngine_Insert(t *testing.T) {
	//engine.Schema.TableDrop(&redis_orm.SchemaTablesTb{})
	//engine.Schema.TableDrop(&redis_orm.SchemaColumnsTb{})
	//engine.Schema.TableDrop(&redis_orm.SchemaIndexsTb{})
	engine.Schema.TableDrop(models.Faq{})
	engine.Schema.CreateTable(models.Faq{})
	//	//t.Logf("tables:%v", engine.Tables)
	//	//engine.TableTruncate(&models.FaqTb{})
	//	for _, table := range engine.Tables {
	//		if strings.Contains(redis_orm.NeedMapTable, table.Name) {
	//			continue
	//		}
	//		t.Logf("table:%v", *table)
	//		for key, column := range table.ColumnsMap {
	//			t.Logf("column:%s,%v", key, column)
	//		}
	//		for key, index := range table.IndexesMap {
	//			t.Logf("index:%s,%v", key, index)
	//		}
	//	}
	//	ary := make([]interface{}, 0)
	//	faq := &models.Faq{
	//		Title:  "index3",
	//		Unique: 12121223,
	//		Hearts: 1,
	//	}
	//	engine.Schema.CreateTable(faq)
	//	ary = append(ary, faq)
	//	faq = &models.Faq{
	//		Title:  "index8",
	//		Unique: 1561223,
	//		Hearts: 2,
	//	}
	//	ary = append(ary, faq)
	//	affected, err := engine.InsertMulti(ary...)
	//	bys, _ := json.Marshal(faq)
	//	t.Logf("InsertMulti faq:%v,affected:%d,err:%v", string(bys), affected, err)
}

//func TestEngine_ReloadTables(t *testing.T) {
//	tables, err := engine.Schema.ReloadTables()
//	if err == nil {
//		for _, table := range tables {
//			t.Logf("table:%v", *table)
//		}
//	} else {
//		t.Logf("SchemaTables2MapTables err:%v", err)
//	}
//}

//func TestEngine_Get(t *testing.T) {
//	faq := &models.Faq{
//		Id: 6,
//	}
//	has, err := engine.Get(faq)
//	bys, _ := json.Marshal(faq)
//	t.Logf("faq:%v,has:%v,err:%v", string(bys), has, err)
//}

//func TestEngine_GetByCombinedIndex(t *testing.T) {
//	faq := &models.Faq{}
//	has, err := engine.GetByCondition(faq, redis_orm.NewSearchCondition(
//		redis_orm.IndexType_IdScore,
//		"1&index",
//		"1&index",
//		"Type",
//		"Title",
//	))
//
//	bys, _ := json.Marshal(faq)
//	t.Logf("faq:%v,has:%v,err:%v", string(bys), has, err)
//}

//func TestEngine_Find(t *testing.T) {
//	//engine.IndexReBuild(models.Faq{})
//	var faqAry []*models.Faq
//	count, err := engine.Find(0, 30, redis_orm.NewSearchConditionV2(
//		1,
//		10,
//		"Id",
//	), &faqAry)
//	bys, _ := json.Marshal(faqAry)
//	t.Logf("faqAry:%v,count:%v,err:%v", string(bys), count, err)
//
//}

//func TestEngine_Update(t *testing.T) {
//	faq := &models.Faq{
//		Id:    1,
//		Title: "test55",
//	}
//	err := engine.Update(faq, "Title")
//	t.Logf("TestEngine_Update err:%v", err)
//}

//func TestEngine_UpdateMulti(t *testing.T) {
//	faq := &models.Faq{
//		Title: "test51",
//	}
//	affectedRow, err := engine.UpdateMulti(faq, redis_orm.NewSearchCondition(
//		redis_orm.IndexType_IdScore,
//		"1",
//		"1",
//		"Id",
//	),
//		"Title")
//	t.Logf("TestEneine_UpdateMulti affectedRow:%d,err:%v", affectedRow, err)
//}

//func TestEngine_Delete(t *testing.T) {
//	faq := &models.Faq{
//		Id: 6,
//	}
//	err := engine.Delete(faq)
//	t.Logf("Delete faq:%v, err:%v", faq, err)
//}

//
//func TestEngine_TableDrop(t *testing.T) {
//	faq := &models.Faq{}
//	err := engine.Schema.TableDrop(faq)
//	t.Logf("TestEngine_TableDrop err:%v", err)
//}
//func TestEngine_DeleteMulti(t *testing.T) {
//	faq := &models.Faq{
//		Title: "test51",
//	}
//	affectedRow, err := engine.DeleteByCondition(faq, redis_orm.NewSearchCondition(
//		redis_orm.IndexType_IdScore,
//		"1",
//		"4",
//		"Id",
//	),
//		"Title")
//	t.Logf("TestEngine_DeleteMulti affectedRow:%d,err:%v", affectedRow, err)
//}
