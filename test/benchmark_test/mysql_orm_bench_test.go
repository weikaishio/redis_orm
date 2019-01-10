package benchmark

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/mkideal/log"
	"github.com/weikaishio/redis_orm/example/models"
)

var (
	orm *xorm.Engine
)

func init() {
	driver := "mysql"
	host := "127.0.0.1:3306"
	database := "bg_db"
	username := "root"
	password := ""
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&&allowOldPasswords=1&parseTime=true", username, password, host, database)

	var err error
	orm, err = xorm.NewEngine(driver, dataSourceName)
	if err != nil {
		log.Fatal("NewEngine:%s,err:%v", dataSourceName, err)
	}
}

func Benchmark_MysqlOrmGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		faq := &models.FaqTb{
			Id: int64(i),
		}
		orm.Get(faq)
	}
}
func Benchmark_MysqlOrmInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		faq := &models.FaqTb{
			Title:   fmt.Sprintf("title%d", i),
			Content: fmt.Sprintf("contente%d", i),
			Hearts:  i,
		}
		orm.InsertOne(faq)
	}
}
