package redis_orm

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

//todo: SearchCondition not a elegant way..
func (e *Engine) GetByCondition(bean interface{}, searchCon *SearchCondition) (bool, error) {
	table, err := e.GetTable(bean)
	if err != nil {
		return false, err
	}

	getId, err := e.indexGetId(searchCon, bean)
	if err != nil {
		return false, err
	}
	if getId == 0 {
		return false, Err_DataNotAvailable
	}

	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	colValue := reflectVal.FieldByName(table.PrimaryKey)
	colValue.SetInt(getId)

	return e.Get(bean)
}
func (e *Engine) Get(bean interface{}) (bool, error) {
	table, err := e.GetTable(bean)
	if err != nil {
		return false, err
	}
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return false, Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	getId, err := e.indexGetId(&SearchCondition{
		SearchColumn:  table.PrimaryKey,
		IndexType:     IndexType_IdMember,
		FieldMinValue: pkInt,
		FieldMaxValue: pkInt,
	}, bean)
	if err != nil {
		return false, err
	}
	if getId == 0 {
		return false, nil
	}

	fields := make([]string, 0)
	for _, colName := range table.ColumnsSeq {
		fieldName := GetFieldName(pkInt, colName)
		fields = append(fields, fieldName)
	}
	valAry, err := e.redisClient.HMGet(table.GetTableKey(), fields...).Result()
	if err != nil {
		return false, err
	} else if valAry == nil {
		return false, nil
	}
	if len(fields) != len(valAry) {
		return false, Err_FieldValueInvalid
	}
	fmt.Printf("valAry:%v", valAry)
	//todo: more safe assignment way？
	for i, val := range valAry {
		if val == nil && table.ColumnsSeq[i] == table.PrimaryKey {
			return false, nil
		}
		if val == nil {
			continue
		}
		colValue := reflectVal.FieldByName(table.ColumnsSeq[i])
		switch colValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8:
			var valInt int64
			SetInt64FromStr(&valInt, val.(string))
			colValue.SetInt(valInt)
		case reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			var valInt uint64
			SetUint64FromStr(&valInt, val.(string))
			colValue.SetUint(valInt)
		case reflect.String:
			colValue.SetString(val.(string))
		default:
			colValue.Set(reflect.ValueOf(val))
		}
	}

	return true, nil
}

func (e *Engine) Find(limit, offset int64, searchCon *SearchCondition, beanAry []interface{}) (int64, error) {

	return 0, nil
}

//todo:unique index is exist?
func (e *Engine) Insert(bean interface{}) error {
	table, err := e.GetTable(bean)
	if err != nil {
		return err
	}
	var lastId int64
	if table.AutoIncrement != "" {
		lastId, err = e.redisClient.HIncrBy(table.GetTableKey(), table.GetAutoIncrKey(), 1).Result()
		if err != nil {
			return err
		}
	}

	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	valMap := make(map[string]interface{})

	for colName, col := range table.ColumnsMap {
		fieldName := GetFieldName(lastId, colName)
		if col.IsAutoIncrement {
			valMap[fieldName] = ToString(lastId)
			colValue := reflectVal.FieldByName(colName)
			colValue.SetInt(lastId)
		} else if col.IsUpdated {

		} else if col.IsCreated {
			valMap[fieldName] = time.Now().In(e.TZLocation).Unix()
		} else {
			colValue := reflectVal.FieldByName(colName)
			valMap[fieldName] = ToString(colValue.Interface())
			if valMap[fieldName] == "" && col.DefaultValue != "" {
				valMap[fieldName] = col.DefaultValue
			}
		}
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		e.indexUpdate(bean)
	}
	return nil
}
func (e *Engine) Update(bean interface{}, cols ...string) error {
	table, err := e.GetTable(bean)
	if err != nil {
		return err
	}
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	getId, err := e.indexGetId(&SearchCondition{
		SearchColumn:  table.PrimaryKey,
		IndexType:     IndexType_IdMember,
		FieldMinValue: pkInt,
		FieldMaxValue: pkInt,
	}, bean)
	if err != nil {
		return err
	}
	if getId == 0 {
		return Err_DataNotAvailable
	}

	valMap := make(map[string]interface{})

	for colName, col := range table.ColumnsMap {
		fieldName := GetFieldName(pkInt, colName)
		for _, colUpdate := range cols {
			if strings.ToLower(colUpdate) != strings.ToLower(colName) {
				continue
			} else if col.IsCreated {
				continue
			}
			colValue := reflectVal.FieldByName(colName)
			if colValue.IsValid() {
				valMap[fieldName] = ToString(colValue.Interface())
			} else {
				valMap[fieldName] = col.DefaultValue
			}
		}
		if col.IsUpdated {
			valMap[fieldName] = time.Now().In(e.TZLocation).Unix()
			continue
		}
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		e.indexUpdate(bean)
	}
	return nil
}
func (e *Engine) Delete(bean interface{}) error {
	table, err := e.GetTable(bean)
	if err != nil {
		return err
	}
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	fields := make([]string, 0)
	for _, colName := range table.ColumnsSeq {
		fieldName := GetFieldName(pkInt, colName)
		fields = append(fields, fieldName)
	}

	_, err = e.redisClient.HDel(table.GetTableKey(), fields...).Result()
	if err == nil {
		e.indexDelete(bean)
	}
	return nil
}
