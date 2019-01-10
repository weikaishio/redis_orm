package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/mkideal/log"
	"github.com/weikaishio/redis_orm"
	"github.com/weikaishio/redis_orm/example/models"
	"github.com/weikaishio/redis_orm/sync2db"
	"time"
)

var (
	engine *redis_orm.Engine

	mysqlOrm *xorm.Engine
)

func init() {
	driver := "mysql"
	host := "127.0.0.1:3306"
	database := "bg_db"
	username := "root"
	password := ""
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&&allowOldPasswords=1&parseTime=true", username, password, host, database)

	var err error
	mysqlOrm, err = xorm.NewEngine(driver, dataSourceName)
	if err != nil {
		panic(fmt.Sprintf("xorm.NewEngine:%s,err:%v\n", dataSourceName, err))
	}
	mysqlOrm.ShowSQL(true)

	options := redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	}

	redisClient := redis.NewClient(&options)

	engine = redis_orm.NewEngine(redisClient)
	engine.IsShowLog(true)
	_, err = engine.Schema.ReloadTables()
	if err != nil {
		panic(fmt.Sprintf("ReloadTables err:%v", err))
	}
	engine.SetSync2DB(mysqlOrm)
}
func main() {
	time.Sleep(time.Second)
	faq := &models.FaqTb{
		Title:  "index112abc1234456",
		Hearts: 11111,
	}
	log.Trace("test log")
	//engine.Insert(faq)
	//engine.Schema.TableDrop(faq)
	//engine.Schema.CreateTable(faq)
	ary := make([]interface{}, 0)
	ary = append(ary, faq)
	affected, err := engine.InsertMulti(ary...)
	engine.Printfln("InsertMulti(%v),affected:%d, err:%v", ary, affected, err)

	//engine.UpdateMulti(faq, redis_orm.NewSearchConditionV2(1, 14, "Id"), "Hearts")

	val, err := engine.Incr(&models.FaqTb{Id: 2}, "Hearts", 333)
	engine.Printfln("engine.Incr val:%d,err:%v", val, err)

	sync2db.ListenQuitAndDump()
	engine.Quit()
	engine.Printfln("quit")
}
