package redis_orm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/mkideal/log"
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

*/

type Engine struct {
	redisClient *redis.Client
	Tables      map[reflect.Type]*Table
	tablesMutex *sync.RWMutex

	showExecTime bool
	TZLocation   *time.Location // The timezone of the application
}

func NewEngine(redisCli *redis.Client) *Engine {
	return &Engine{
		redisClient: redisCli,
		Tables:      make(map[reflect.Type]*Table),
		tablesMutex: &sync.RWMutex{},
		TZLocation:  time.Local,
	}
}

func (e *Engine) GetTable(bean interface{}) (*Table, error) {
	beanValue := reflect.ValueOf(bean)
	if beanValue.Kind() != reflect.Ptr {
		return nil, errors.New("needs a pointer to a value")
	} else if beanValue.Elem().Kind() == reflect.Ptr {
		return nil, errors.New("a pointer to a pointer is not allowed")
	}

	if beanValue.Elem().Kind() == reflect.Struct {
		e.tablesMutex.RLock()
		table, ok := e.Tables[beanValue.Type()]
		e.tablesMutex.RUnlock()
		if !ok {
			var err error
			table, err = e.mapTable(reflect.Indirect(beanValue))
			if err != nil {
				return nil, err
			}
			e.tablesMutex.Lock()
			e.Tables[beanValue.Type()] = table
			e.tablesMutex.Unlock()
		}

		return table, nil
	}
	return nil, errors.New("not support kind")
}
func GetFieldName(pkId int64, colName string) string {
	return fmt.Sprintf("%d_%s", pkId, colName)
}
func (e *Engine) mapTable(v reflect.Value) (*Table, error) {
	typ := v.Type()
	table := NewEmptyTable()
	table.Type = typ
	table.Name = strings.ToLower(e.tbName(v))
	//ptr or struct:typ.NumField()
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag
		rdsTagStr := tag.Get(TagIdentifier)
		//fmt.Printf("%d rdsTagStr:%s\n", i, rdsTagStr)

		var col *Column
		fieldValue := v.Field(i)
		fieldType := fieldValue.Type()

		if rdsTagStr != "" {
			col = NewEmptyColumn(typ.Field(i).Name)
			tags := splitTag(rdsTagStr)
			for j, key := range tags {
				keyLower := strings.ToLower(key)
				if keyLower == TagIndex {
					table.AddIndex(fieldType, col.Name)
				} else if keyLower == TagDefaultValue {
					if len(tags) > j {
						col.DefaultValue = strings.Trim(tags[j+1], "'")
					}
				} else if keyLower == TagPrimaryKey {
					col.IsPrimaryKey = true
					table.AddIndex(fieldType, col.Name)
				} else if keyLower == TagAutoIncrement {
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
					//todo:combined index
					table.AddIndex(fieldType, col.Name)
					continue
				} else {
					//abondon
				}
			}
			table.AddColumn(col)
		} else {
			log.Warn("MapTable field:%s, not has tag", typ.Field(i).Name)
		}
	}

	bys, _ := json.Marshal(table)
	log.Trace("table:%v", string(bys))
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
func (e *Engine) tbName(v reflect.Value) string {
	return strings.ToLower(v.Type().Name())
}

//del the hashkey, it will del all elements for this hash
func (e *Engine) DropTable(bean interface{}) error {
	return nil
}

//keys tb:*
func (e *Engine) ShowTables() []string {
	return nil
}
