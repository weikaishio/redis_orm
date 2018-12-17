package redis_orm

import (
	"fmt"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/mkideal/log"
	"strings"
)

//todo:combined index
func (e *Engine) indexGetId(searchCon *SearchCondition, bean interface{}) (int64, error) {
	table, err := e.GetTable(bean)
	if err != nil {
		return 0, err
	}
	index, ok := table.IndexesMap[strings.ToLower(searchCon.SearchColumn)]
	if !ok {
		return 0, err
	}
	//beanValue := reflect.ValueOf(bean)
	//reflectVal := reflect.Indirect(beanValue)

	//colValue := reflectVal.FieldByName(colName)
	switch index.Type {
	case IndexType_IdMember:
		res, err := e.redisClient.ZRangeByScore(index.NameKey, redis.ZRangeBy{
			Min:    ToString(searchCon.FieldMinValue),
			Max:    ToString(searchCon.FieldMaxValue),
			Offset: 0,
			Count:  1,
		}).Result()
		if err == nil {
			if len(res) == 0 {
				return 0, nil
			}
			var id int64
			SetInt64FromStr(&id, res[0])
			return id, nil
		} else {
			log.Error("indexGetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
		}
	case IndexType_IdScore:
		res, err := e.redisClient.ZScore(index.NameKey, searchCon.FieldMinValue.(string)).Result()
		if err == nil {
			return int64(res), nil
		} else {
			log.Error("indexGetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
		}
	case IndexType_UnSupport:
		log.Error("indexGetId unsupport index type")
	}
	return 0, nil
}

func (e *Engine) indexDelete(bean interface{}) error {
	table, err := e.GetTable(bean)
	if err != nil {
		return err
	}

	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	typ := reflectVal.Type()
	_, has := typ.FieldByName(table.PrimaryKey)
	if !has {
		return Err_PrimaryKeyNotFound
	}
	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}
	//fmt.Printf("pkFieldValue:%v,int:%d\n", pkFieldValue, pkFieldValue.Int())
	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		fieldValue := reflectVal.FieldByName(index.ColumnName[0])
		if err != nil {
			log.Error("indexDelete GetFieldValue(%s) err:%v", index.ColumnName, err)
			return err
		}
		fmt.Printf("indexDelete k:%v, index:%v, fieldValue:%v\n", index.ColumnName, index, fieldValue)
		switch index.Type {
		case IndexType_IdMember:
			log.Trace("indexDelete %s:%v", index.NameKey, pkFieldValue.Int())
			_, err := e.redisClient.ZRem(index.NameKey, pkFieldValue.Int()).Result()
			if err != nil {
				log.Warn("indexDelete %s:%v,err:%v", index.NameKey, pkFieldValue.Int(), err)
			}
		case IndexType_IdScore:
			log.Trace("indexDelete %s:%v", index.NameKey, fieldValue.Interface())
			_, err := e.redisClient.ZRem(index.NameKey, fieldValue.Interface()).Result()
			if err != nil {
				log.Warn("indexDelete %s:%v,err:%v", index.NameKey, fieldValue.Interface(), err)
			}
		case IndexType_UnSupport:
			log.Error("indexDelete unsupport index type")
		}
	}
	return nil
}

//todo: no thread safety! watch?
func (e *Engine) indexUpdate(bean interface{}) error {
	table, err := e.GetTable(bean)
	if err != nil {
		return err
	}

	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	typ := reflectVal.Type()
	//fmt.Printf("table.PrimaryKey:%s\n", table.PrimaryKey)
	_, has := typ.FieldByName(table.PrimaryKey)
	if !has {
		return Err_PrimaryKeyNotFound
	}
	//fmt.Printf("IndexUpdate pkField:%v\n", pkField)
	//pkValue := reflectVal.FieldByName(pkField.Name)
	//fmt.Printf("pkValue.String():%s,%v,%v\n", pkValue.String(), pkValue.Type(), pkValue.Int())
	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}
	fmt.Printf("pkFieldValue:%v,int:%d\n", pkFieldValue, pkFieldValue.Int())
	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		fieldValue := reflectVal.FieldByName(index.ColumnName[0])
		if err != nil {
			log.Error("IndexUpdate GetFieldValue(%s) err:%v", index.ColumnName, err)
			return err
		}
		fmt.Printf("IndexUpdate k:%v, index:%v, fieldValue:%v\n", index.ColumnName, index, fieldValue)
		switch index.Type {
		case IndexType_IdMember:
			var score float64
			switch fieldValue.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8:
				fallthrough
			case reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				score = float64(fieldValue.Int())
			case reflect.String:
				SetFloat64FromStr(&score, fieldValue.String())
			default:
				SetFloat64FromStr(&score, fmt.Sprintf("%v", fieldValue.Interface()))
			}
			redisZ := redis.Z{
				Member: pkFieldValue.Int(),
				Score:  score,
			}
			log.Trace("IndexUpdate %s:%v", index.NameKey, redisZ)
			_, err = e.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				log.Warn("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
			}
		case IndexType_IdScore:
			redisZ := redis.Z{
				Member: fieldValue.Interface(),
				Score:  float64(pkFieldValue.Int()),
			}
			log.Trace("IndexUpdate %s:%v", index.NameKey, redisZ)
			_, err = e.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				log.Warn("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
			}
		case IndexType_UnSupport:
			log.Error("IndexUpdate unsupport index type")
		}
	}
	return nil
}
