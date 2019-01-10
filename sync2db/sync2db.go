package sync2db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/weikaishio/distributed_lib/db_lazy"
	"reflect"
	"sync"
	"time"
)

type Sync2DB struct {
	mysqlOrm  *xorm.Engine
	isShowLog bool
	*db_lazy.LazyMysql
	wait *sync.WaitGroup
}

func (s *Sync2DB) IsShowLog(isShow bool) {
	s.isShowLog = isShow
}
func NewSync2DB(mysqlOrm *xorm.Engine, wait *sync.WaitGroup) *Sync2DB {
	sync2DB := &Sync2DB{
		mysqlOrm: mysqlOrm,
		wait:     wait,
	}
	sync2DB.LazyMysql = db_lazy.NewLazyMysql(mysqlOrm, 10)
	go func() {
		go sync2DB.LazyMysql.Exec()
		ListenQuitAndDump() //expose a quit method or listen kill process signal
		sync2DB.LazyMysql.Quit()
		sync2DB.wait.Done()
	}()
	return sync2DB
}
func (s *Sync2DB) Create2DB(bean interface{}) error {
	err := s.mysqlOrm.Sync(bean)
	if err != nil {
		s.Printfln("mysqlOrm.Sync(%v),err:%v", reflect.TypeOf(bean).Name(), err)
	} else {
		s.Printfln("mysqlOrm.Sync(%v)", reflect.TypeOf(bean).Name())
	}
	return err
}
func (s *Sync2DB) Printfln(format string, a ...interface{}) {
	if s.isShowLog {
		s.Printf(format, a...)
		fmt.Print("\n")
	}
}

func (s *Sync2DB) Printf(format string, a ...interface{}) {
	if s.isShowLog {
		fmt.Printf(fmt.Sprintf("[redis_orm %s] : %s", time.Now().Format("06-01-02 15:04:05"), format), a...)
	}
}

//首次初始化
func (s *Sync2DB) Sync() {
}

//func (s *Sync2DB) ToMysql(offset, limit int64, searchCon *redis_orm.SearchCondition, beanAry interface{}) error {
//	count, err := s.engine.Find(offset, limit, searchCon, beanAry)
//	if err != nil {
//		return err
//	}
//	if count == 0 {
//		return redis_orm.Err_DataNotAvailable
//	}
//	//switch v := beanAry.(type) {
//	//case []interface{}:
//	//	affectedRows, err := s.mysqlOrm.Insert(v...)
//	//	s.engine.Printfln("Insert affectedRows:%d,err:%v", affectedRows, err)
//	//	return err
//	//case *[]interface{}:
//	//	for _, bean := range *v {
//	//		affectedRows, err := s.mysqlOrm.Insert(bean)
//	//		s.engine.Printfln("Insert affectedRows:%d,err:%v", affectedRows, err)
//	//	}
//	//default:
//	sliceValue := reflect.Indirect(reflect.ValueOf(beanAry))
//	sliceElementType := sliceValue.Type().Elem()
//	s.engine.Printfln("sliceValue.Interface():%v\nsliceValue:%v\nsliceElementType:%v", sliceValue.Interface(), sliceValue,sliceElementType)
//	//sliceElementType := sliceValue.Type().Elem()
//
//	for _, bean := range sliceValue.Interface().(reflect.TypeOf(sliceValue)) {
//		affectedRows, err := s.mysqlOrm.Insert(bean)
//		s.engine.Printfln("Insert affectedRows:%d,err:%v", affectedRows, err)
//	}
//	//s.engine.Printfln("ToMysql type:%v err", reflect.TypeOf(v))
//	//}
//	return nil
//}
