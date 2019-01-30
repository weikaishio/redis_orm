package redis_orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/weikaishio/redis_orm/sync2db"
)

//type ORM interface {
//	Insert()
//	Update()
//	Delete()
//	Find()
//	Get()
//}
/*
从tag获取
索引, 只支持数值 字符 唯一索引的
自增
默认值

todo:链式查询

todo:权限控制~

*/
/*
todo:concurrency safe
并发的处理
1、唯一值的写入
解决方案：
2、自增（HIncrBy,HIncrByFloat）的原子可靠性
3、CAS

todo:pipeline
*/
type Engine struct {
	redisClient *RedisClientProxy
	Tables      map[string]*Table
	tablesMutex *sync.RWMutex
	isShowLog   bool

	showExecTime bool
	TZLocation   *time.Location // The timezone of the application

	Schema *SchemaEngine
	Index  *IndexEngine

	isSync2DB bool
	syncDB    *sync2db.Sync2DB
	wait      *sync.WaitGroup
}

func NewEngine(redisCli *redis.Client) *Engine {
	redisCliProxy := NewRedisCliProxy(redisCli)
	engine := &Engine{
		redisClient: redisCliProxy,
		Tables:      make(map[string]*Table),
		tablesMutex: &sync.RWMutex{},
		TZLocation:  time.Local,
		isShowLog:   false,
		wait:        &sync.WaitGroup{},
	}
	redisCliProxy.engine = engine

	index := NewIndexEngine(engine)
	engine.Index = index

	schema := NewSchemaEngine(engine)
	engine.Schema = schema

	engine.Schema.ReloadTables()
	return engine
}
func (e *Engine) SetSync2DB(mysqlOrm *xorm.Engine, lazyTimeSecond int) {
	e.syncDB = sync2db.NewSync2DB(mysqlOrm, lazyTimeSecond, e.wait)
	e.syncDB.IsShowLog(e.isShowLog)
	e.isSync2DB = true
	e.wait.Add(1)
}
func (e *Engine) IsShowLog(isShow bool) {
	e.isShowLog = isShow
}
func (e *Engine) Quit() {
	e.Printfln("redis_orm Engine's Quit begin")
	e.wait.Wait()
	e.Printfln("redis_orm Engine's Quit end")
}
func (e *Engine) Printfln(format string, a ...interface{}) {
	if e.isShowLog {
		e.Printf(format, a...)
		fmt.Print("\n")
	}
}

//todo: command->redis client
func (e *Engine) Printf(format string, a ...interface{}) {
	if e.isShowLog {
		fmt.Printf(fmt.Sprintf("[redis_orm %s] : %s", time.Now().Format("06-01-02 15:04:05"), format), a...)
	}
}
func (e *Engine) GetTableByName(tableName string) (*Table, bool) {
	e.tablesMutex.RLock()
	defer e.tablesMutex.RUnlock()
	table, has := e.Tables[tableName]
	if !has {
		e.Printfln("GetTableByName(%s) !has", tableName)
	}
	return table, has
}
func (e *Engine) GetTableByReflect(beanValue, beanIndirectValue reflect.Value) (*Table, error) {
	if beanValue.Kind() != reflect.Ptr {
		return nil, Err_NeedPointer
	} else if beanValue.Elem().Kind() == reflect.Ptr {
		return nil, Err_NotSupportPointer2Pointer
	}

	if beanValue.Elem().Kind() == reflect.Struct ||
		beanValue.Elem().Kind() == reflect.Interface {
		e.tablesMutex.RLock()
		table, ok := e.Tables[e.TableName(beanIndirectValue)]
		e.tablesMutex.RUnlock()
		if !ok {
			if strings.Contains(NeedMapTable, e.TableName(beanIndirectValue)) {
				var err error
				table, err = e.mapTable(beanIndirectValue)
				if err != nil {
					return nil, err
				}
				e.tablesMutex.Lock()
				e.Tables[table.Name] = table
				e.tablesMutex.Unlock()
			} else {
				e.Printfln("table:%s, not found", e.TableName(beanIndirectValue))
				return nil, ERR_UnKnowTable
			}
		}

		return table, nil
	}
	return nil, Err_UnSupportedType
}
func GetFieldName(pkId interface{}, colName string) string {
	return fmt.Sprintf("%v_%s", pkId, colName)
}
func MapTableColumnFromTag(table *Table, seq int, columnName string, columnType reflect.Kind, rdsTagStr string) error {
	col := NewEmptyColumn(columnName)
	col.Seq = byte(seq)
	col.DataType = columnType.String()
	tags := splitTag(rdsTagStr)
	var (
		isIndex   bool
		isUnique  bool
		indexName string
	)
	for j, key := range tags {
		keyLower := strings.ToLower(key)
		if keyLower == TagIndex {
			isIndex = true
			indexName = col.Name
		} else if keyLower == TagUniqueIndex {
			isIndex = true
			indexName = col.Name
			isUnique = true
		} else if keyLower == TagDefaultValue {
			if len(tags) > j {
				col.DefaultValue = strings.Trim(tags[j+1], "'")
			}
		} else if keyLower == TagPrimaryKey {
			col.IsPrimaryKey = true
			isIndex = true
			indexName = col.Name
		} else if keyLower == TagAutoIncrement {
			if table.AutoIncrement != "" {
				return Err_MoreThanOneIncrementColumn
			}
			col.IsAutoIncrement = true
		} else if keyLower == TagComment {
			if len(tags) > j {
				col.Comment = strings.Trim(tags[j+1], "'")
			}
		} else if keyLower == TagCreatedAt {
			col.IsCreated = true
		} else if keyLower == TagUpdatedAt {
			col.IsUpdated = true
		} else if keyLower == TagCombinedindex {
			//Done:combined index
			if columnType != reflect.String && columnType != reflect.Int64 {
				return Err_CombinedIndexTypeError
			}
			if len(tags) > j && tags[j+1] != "" {
				isIndex = true
				indexName = tags[j+1]
				col.IsCombinedIndex = true
			}
			continue
		} else if keyLower == TagSync2DB {
			table.IsSync2DB = true
		} else {
			//abondon
		}
	}
	table.AddColumn(col)
	if isIndex {
		table.AddIndex(columnType, indexName, col.Name, col.Comment, isUnique, col.Seq)
	}
	return nil
}
func (e *Engine) mapTable(v reflect.Value) (*Table, error) {
	typ := v.Type()
	table := NewEmptyTable()
	//table.Type = typ
	table.Name = strings.ToLower(e.TableName(v))
	table.Comment = Camel2Underline(e.TableName(v))
	//ptr or struct:typ.NumField()
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag
		rdsTagStr := tag.Get(TagIdentifier)

		//var col *Column
		fieldValue := v.Field(i)
		fieldType := fieldValue.Type()

		if rdsTagStr != "" {
			err := MapTableColumnFromTag(table, i, typ.Field(i).Name, fieldType.Kind(), rdsTagStr)
			if err != nil {
				return table, err
			}
			//col = NewEmptyColumn(typ.Field(i).Name)
			//col.Seq = byte(i)
			//col.DataType = fieldType.Kind().String()
			//e.Printfln("col.DataType:%s", col.DataType)
			//tags := splitTag(rdsTagStr)
			//var (
			//	isIndex   bool
			//	isUnique  bool
			//	indexName string
			//)
			//for j, key := range tags {
			//	keyLower := strings.ToLower(key)
			//	if keyLower == TagIndex {
			//		isIndex = true
			//		indexName = col.Name
			//	} else if keyLower == TagUniqueIndex {
			//		isIndex = true
			//		indexName = col.Name
			//		isUnique = true
			//	} else if keyLower == TagDefaultValue {
			//		if len(tags) > j {
			//			col.DefaultValue = strings.Trim(tags[j+1], "'")
			//		}
			//	} else if keyLower == TagPrimaryKey {
			//		col.IsPrimaryKey = true
			//		isIndex = true
			//		indexName = col.Name
			//	} else if keyLower == TagAutoIncrement {
			//		if table.AutoIncrement != "" {
			//			return nil, Err_MoreThanOneIncrementColumn
			//		}
			//		col.IsAutoIncrement = true
			//	} else if keyLower == TagComment {
			//		if len(tags) > j {
			//			col.Comment = strings.Trim(tags[j+1], "'")
			//		}
			//	} else if keyLower == TagCreatedAt {
			//		col.IsCreated = true
			//	} else if keyLower == TagUpdatedAt {
			//		col.IsUpdated = true
			//	} else if keyLower == TagCombinedindex {
			//		//Done:combined index
			//		if fieldType.Kind() != reflect.String && fieldType.Kind() != reflect.Int64 {
			//			return nil, Err_CombinedIndexTypeError
			//		}
			//		if len(tags) > j && tags[j+1] != "" {
			//			isIndex = true
			//			indexName = tags[j+1]
			//			col.IsCombinedIndex = true
			//		}
			//		continue
			//	} else if keyLower == TagSync2DB {
			//		table.IsSync2DB = true
			//	} else {
			//		//abondon
			//	}
			//}
			////col.Type = fieldType
			//table.AddColumn(col)
			//if isIndex {
			//	table.AddIndex(fieldType.Kind(), indexName, col.Name, col.Comment, isUnique, col.Seq)
			//}
		} else {
			e.Printfln("MapTable field:%s, not has tag", typ.Field(i).Name)
		}
	}

	//bys, _ := json.Marshal(table)
	//log.Trace("table:%v", string(bys))
	return table, nil
}
func splitTag(tag string) (tags []string) {
	tag = strings.TrimSpace(tag)
	var hasQuote = false
	var lastIdx = 0
	for i, t := range tag {
		if t == '\'' {
			hasQuote = !hasQuote
		} else if t == ' ' {
			if lastIdx < i && !hasQuote {
				tags = append(tags, strings.TrimSpace(tag[lastIdx:i]))
				lastIdx = i + 1
			}
		}
	}
	if lastIdx < len(tag) {
		tags = append(tags, strings.TrimSpace(tag[lastIdx:]))
	}
	return
}
func (e *Engine) TableName(v reflect.Value) string {
	return strings.ToLower(v.Type().Name())
}

func SetDefaultValue(col *Column, value *reflect.Value) {
	if !value.CanSet() {
		return
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() == 0 {
			var valInt int64
			SetInt64FromStr(&valInt, col.DefaultValue)
			value.SetInt(valInt)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Uint() == 0 {
			var valInt uint64
			SetUint64FromStr(&valInt, col.DefaultValue)
			value.SetUint(valInt)
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() == 0 {
			var valInt float64
			SetFloat64FromStr(&valInt, col.DefaultValue)
			value.SetFloat(valInt)
		}
	case reflect.String:
		if ToString(value.Interface()) == "" {
			value.SetString(col.DefaultValue)
		}
	case reflect.Map:
		//todo:SetValue4Map
	case reflect.Bool:
		if !value.Bool() {
			value.SetBool(false)
		}
	default:
	}
}
func SetValue(val interface{}, value *reflect.Value) {
	if !value.CanSet() {
		return
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var valInt int64
		SetInt64FromStr(&valInt, ToString(val))
		value.SetInt(valInt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var valInt uint64
		SetUint64FromStr(&valInt, ToString(val))
		value.SetUint(valInt)
	case reflect.Float32, reflect.Float64:
		var valInt float64
		SetFloat64FromStr(&valInt, ToString(val))
		value.SetFloat(valInt)
	case reflect.String:
		value.SetString(ToString(val))
	case reflect.Map:
		//todo:SetValue4Map
		//reflect.TypeOf()
		//value.Set(reflect.MapOf())
	case reflect.Bool:
		var valBool bool
		SetBoolFromStr(&valBool, ToString(val))
		value.SetBool(valBool)
	default:
		value.Set(reflect.ValueOf(val))
	}
}
