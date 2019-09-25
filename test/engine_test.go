package test

import (
	"encoding/json"
	"testing"

	"github.com/go-redis/redis"
	"github.com/weikaishio/redis_orm"
	"github.com/weikaishio/redis_orm/example/models"
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
	//_, err := engine.Schema.ReloadTables()
	//if err != nil {
	//	engine.Printfln("ReloadTables err:%v", err)
	//}
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

func TestEngine_Insert(t *testing.T) {
	//engine.Schema.TableDrop(&redis_orm.SchemaTablesTb{})
	//engine.Schema.TableDrop(&redis_orm.SchemaColumnsTb{})
	//engine.Schema.TableDrop(&redis_orm.SchemaIndexsTb{})
	//engine.Schema.CreateTable(models.FaqTb{})
	//engine.Schema.TableDrop(&models.FaqTestTb{})
	//engine.Schema.TableTruncate(&models.FaqTestTb{})
	//err := engine.Schema.CreateTable(&models.FaqTestTb{})
	//if err != nil {
	//	t.Logf("CreateTable err:%v", err)
	//}
	err := engine.Insert(&models.FaqTb{Type: 101, Title: "title3", Content: "content"})
	if err != nil {
		errWithCode:=err.(*redis_orm.ErrorWithCode)
		t.Logf("Insert err:%v,%v", errWithCode.Code(),errWithCode.Error())
	}
	//t.Logf("tables:%v", engine.Tables)
	//engine.TableTruncate(&models.FaqTb{})
	//for _, table := range engine.Tables {
	//	if strings.Contains(redis_orm.NeedMapTable, table.Name) {
	//		continue
	//	}
	//	t.Logf("table:%v", *table)
	//	for key, column := range table.ColumnsMap {
	//		t.Logf("column:%s,%v", key, column)
	//	}
	//	for key, index := range table.IndexesMap {
	//		t.Logf("index:%s,%v", key, index)
	//	}
	//}
	//ary := make([]interface{}, 0)
	//faq := models.Faq{
	//	Title:  "index3",
	//	Unique: 3,
	//	Hearts: 1,
	//}
	//engine.Schema.CreateTable(faq)
	//ary = append(ary, faq)
	//faq = models.Faq{
	//	Title:  "index4",
	//	Unique: 4,
	//	Hearts: 2,
	//}
	//ary = append(ary, faq)
	////for i := 4000; i < 10000; i++ {
	////	faq = &models.Faq{
	////		Title:  fmt.Sprintf("index%d", i),
	////		Unique: int64(i),
	////		Hearts: i,
	////	}
	////	ary = append(ary, faq)
	////}
	//affected, err := engine.InsertMulti(ary...)
	//bys, _ := json.Marshal(faq)
	//t.Logf("InsertMulti faq:%v,affected:%d,err:%v", string(bys), affected, err)
}

//func TestEngine_Incr(t *testing.T) {
//	faq := models.Faq{
//		Id: 2,
//	}
//	engine.Schema.CreateTable(faq)
//	val, err := engine.Incr(faq, "Hearts", 1)
//	t.Logf("val:%v,err:%v", val, err)
//}

//func TestEngine_Query(t *testing.T) {
//	tableName := "faq"
//	table, has := engine.Schema.GetTableByName(tableName)
//	if !has {
//		t.Logf("GetTableByName(%s),has:%v", tableName, has)
//		return
//	}
//	resAry, count, err := engine.Query(0, 210, redis_orm.NewSearchConditionV2(9889,
//		9889,
//		table.PrimaryKey), table,"Id","Content")
//	t.Logf("resAry:%v,count:%d,err:%v", resAry, count, err)
//}

//func TestEngine_Query(t *testing.T) {
//	tableName := "faq"
//	table, has := engine.Schema.GetTableByName(tableName)
//	if !has {
//		t.Logf("GetTableByName(%s),has:%v", tableName, has)
//		return
//	}
//	resAry, count, err := engine.Query(0, 10, redis_orm.NewSearchConditionV2(redis_orm.ScoreMin,
//		redis_orm.ScoreMax,
//		table.PrimaryKey), table, "Id", "Title", "test", "CreatedAt", "Content", "Hearts")
//	t.Logf("resAry:%v,count:%d,err:%v", resAry, count, err)
//}

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

func TestEngine_Get(t *testing.T) {
	faq := &models.FaqTb{}

	searchCon := redis_orm.NewSearchConditionV2(101, 101, "Type")
	has, err := engine.GetByCondition(faq, searchCon)
	bys, _ := json.Marshal(faq)
	t.Logf("faq:%v,has:%v,err:%v", string(bys), has, err)

	//err := engine.GetDefaultValue(faq)
	//bys, _ := json.Marshal(faq)
	//t.Logf("faq:%v,has:%v,err:%v", string(bys), 0, err)
}

//func TestEngine_GetByCombinedIndex(t *testing.T) {
//	faq := models.Faq{}
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

func TestEngine_Find(t *testing.T) {
	//engine.IndexReBuild(models.Faq{})
	var faqAry []models.Faq
	count, err := engine.Find(0, 30, redis_orm.NewSearchConditionV2(
		1,
		10,
		"Id",
	), &faqAry)
	table := engine.Tables["faq"]
	if table != nil {
		t.Logf("table:%v", table.ColumnsMap)
	}
	bys, _ := json.Marshal(faqAry)
	t.Logf("faqAry:%v,count:%v,err:%v", string(bys), count, err)

}

func TestEngine_Update(t *testing.T) {
	faq := models.FaqTb{
		Id:   3,
		Type: 103,
	}
	err := engine.Update(&faq, "Type")
	t.Logf("TestEngine_Update err:%v", err)
}

//
//func TestEngine_UpdateMulti(t *testing.T) {
//	faq := models.Faq{
//		Title: "test51",
//	}
//	affectedRow, err := engine.UpdateMulti(faq, redis_orm.NewSearchCondition(
//		redis_orm.IndexType_IdScore,
//		"1",
//		"100",
//		"Id",
//	),
//		"Hearts")
//
//	t.Logf("TestEngine_UpdateMulti affectedRow:%d,err:%v", affectedRow, err)
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
//	beanValue := reflect.ValueOf(faq)
//	beanIndirectValue := reflect.Indirect(beanValue)
//
//	table, has := engine.GetTableByName(engine.TableName(beanIndirectValue))
//	if !has {
//		t.Logf("GetTableByName !has")
//		return
//	}
//	err := engine.Schema.TableDrop(table)
//	t.Logf("TestEngine_TableDrop err:%v", err)
//}

//func TestEngine_DeleteMulti(t *testing.T) {
//	faq := &models.Faq{
//		Title: "test51",
//	}
//	affectedRow, err := engine.DeleteByCondition(faq, redis_orm.NewSearchConditionV2(
//		"1",
//		"400",
//		"Id",
//	))
//	t.Logf("TestEngine_DeleteMulti affectedRow:%d,err:%v", affectedRow, err)
//}
