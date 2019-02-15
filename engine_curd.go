package redis_orm

import (
	"fmt"
	"github.com/weikaishio/distributed_lib/db_lazy"
	"reflect"
	"time"
)

/*
only support one searchCondition to get or find
todo: SearchCondition not a elegant way..
*/
func (e *Engine) GetByCondition(bean interface{}, searchCon *SearchCondition) (bool, error) {
	beanValue := reflect.ValueOf(bean)
	if beanValue.Kind() != reflect.Ptr {
		return false, Err_NeedPointer
	}
	reflectVal := reflect.Indirect(beanValue)
	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return false, ERR_UnKnowTable
	}

	getId, err := e.Index.GetId(table, searchCon)
	if err != nil {
		return false, err
	}
	if getId == 0 {
		return false, nil
	}
	colValue := reflectVal.FieldByName(table.PrimaryKey)
	if colValue.CanSet() {
		colValue.SetInt(getId)
	} else {
		return false, Err_NeedPointer
	}

	fields := make([]string, 0)
	for _, colName := range table.ColumnsSeq {
		fieldName := GetFieldName(getId, colName)
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
	//todo: any other safer assignment way？
	for i, val := range valAry {
		if val == nil && table.ColumnsSeq[i] == table.PrimaryKey {
			return false, nil
		}
		if val == nil {
			continue
		}
		colValue := reflectVal.FieldByName(table.ColumnsSeq[i])

		SetValue(val, &colValue)
	}
	return e.Get(bean)
}
func (e *Engine) Get(bean interface{}) (bool, error) {
	beanValue := reflect.ValueOf(bean)
	if beanValue.Kind() != reflect.Ptr {
		return false, Err_NeedPointer
	}
	reflectVal := reflect.Indirect(beanValue)
	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return false, ERR_UnKnowTable
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return false, Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	//getId, err := e.Index.GetId(table, &SearchCondition{
	//	SearchColumn: []string{table.PrimaryKey},
	//	//IndexType:     IndexType_IdMember,
	//	FieldMinValue: pkInt,
	//	FieldMaxValue: pkInt,
	//})
	//if err != nil {
	//	return false, err
	//}
	//if getId == 0 {
	//	return false, nil
	//}
	has, err := e.Index.IdIsExist(table, pkInt)
	if err != nil {
		return false, err
	}
	if !has {
		return false, Err_DataNotAvailable
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
	for i, val := range valAry {
		if val == nil && table.ColumnsSeq[i] == table.PrimaryKey {
			return false, nil
		}
		if val == nil {
			continue
		}
		colValue := reflectVal.FieldByName(table.ColumnsSeq[i])

		SetValue(val, &colValue)
	}

	return true, nil
}
func (e *Engine) Count(searchCon *SearchCondition, beanAry interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(beanAry))
	if sliceValue.Kind() != reflect.Slice {
		return 0, Err_NeedPointer
	}

	var (
		table      *Table
		reflectVal reflect.Value
	)
	sliceElementType := sliceValue.Type().Elem()
	if sliceElementType.Kind() == reflect.Ptr {
		if sliceElementType.Elem().Kind() == reflect.Struct {
			beanValue := reflect.New(sliceElementType.Elem())
			reflectVal = reflect.Indirect(beanValue)
			var has bool
			table, has = e.GetTableByName(e.TableName(reflectVal))
			if !has {
				return 0, ERR_UnKnowTable
			}
		}
	} else if sliceElementType.Kind() == reflect.Struct {
		beanValue := reflect.New(sliceElementType)
		reflectVal = reflect.Indirect(beanValue)
		var has bool
		table, has = e.GetTableByName(e.TableName(reflectVal))
		if !has {
			return 0, ERR_UnKnowTable
		}
	}
	if table == nil {
		return 0, Err_UnSupportedTableModel
	}
	return e.count(searchCon, table)
}
func (e *Engine) count(searchCon *SearchCondition, table *Table) (int64, error) {
	count, err := e.Index.Count(table, searchCon)
	return count, err
}
func (e *Engine) TableFromBeanAryReflect(beanAry interface{}) (*Table, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(beanAry))
	if sliceValue.Kind() != reflect.Slice {
		return nil, Err_NeedSlice
	}

	var (
		table      *Table
		reflectVal reflect.Value
	)
	sliceElementType := sliceValue.Type().Elem()
	if sliceElementType.Kind() == reflect.Ptr {
		if sliceElementType.Elem().Kind() == reflect.Struct {
			beanValue := reflect.New(sliceElementType.Elem())
			reflectVal = reflect.Indirect(beanValue)
			var has bool
			table, has = e.GetTableByName(e.TableName(reflectVal))
			if !has {
				return nil, ERR_UnKnowTable
			}
		}
	} else if sliceElementType.Kind() == reflect.Struct || sliceElementType.Kind() == reflect.Interface {
		beanValue := reflect.New(sliceElementType)
		reflectVal = reflect.Indirect(beanValue)
		var has bool
		table, has = e.GetTableByName(e.TableName(reflectVal))
		if !has {
			return nil, ERR_UnKnowTable
		}
	}
	if table == nil {
		return nil, Err_UnSupportedTableModel
	}
	return table, nil
}
func (e *Engine) FileterCols(table *Table, cols ...string) []string {
	if len(cols)==0 {
		return table.ColumnsSeq
	}
filterCols:
	for i, col := range cols {
		_, isExist := table.ColumnsMap[col]
		if !isExist {
			cols = append(cols[:i], cols[i+1:]...)
			goto filterCols
		}
	}
	return cols
}
func (e *Engine) Query(offset, limit int64, condition *SearchCondition, table *Table, cols ...string) ([]map[string]interface{}, int64, error) {
	count, err := e.count(condition, table)
	if err != nil {
		return nil, count, nil
	}
	idAry, err := e.Index.Range(table, condition, offset, limit)
	if err != nil {
		return nil, count, err
	}

	cols = e.FileterCols(table, cols...)

	if len(cols) == 0 {
		cols = table.ColumnsSeq
	}

	fields := make([]string, 0)
	for _, id := range idAry {
		for _, colName := range cols {
			fieldName := GetFieldName(id, colName)
			fields = append(fields, fieldName)
		}
	}
	valAry, err := e.redisClient.HMGet(table.GetTableKey(), fields...).Result()
	if err != nil {
		return nil, count, err
	} else if valAry == nil {
		return nil, count, nil
	}
	if len(fields) != len(valAry) {
		return nil, count, Err_FieldValueInvalid
	}
	var resAry []map[string]interface{}

	for i := 0; i < len(fields); i += len(cols) {
		mapObj := make(map[string]interface{})
		for j, colName := range cols {
			if valAry[i+j] == nil {
				if colName == table.PrimaryKey {
					break
				} else {
					continue
				}
			}
			mapObj[colName] = valAry[i+j]
		}
		resAry = append(resAry, mapObj)
	}
	return resAry, count, nil
}
func (e *Engine) Find(offset, limit int64, searchCon *SearchCondition, beanAry interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(beanAry))
	sliceElementType := sliceValue.Type().Elem()

	table, err := e.TableFromBeanAryReflect(beanAry)
	if err != nil {
		return 0, err
	}
	count, err := e.count(searchCon, table)
	if err != nil {
		return 0, nil
	}
	idAry, err := e.Index.Range(table, searchCon, offset, limit)
	if err != nil {
		return 0, err
	}

	fields := make([]string, 0)
	for _, id := range idAry {
		for _, colName := range table.ColumnsSeq {
			fieldName := GetFieldName(id, colName)
			fields = append(fields, fieldName)
		}
	}
	valAry, err := e.redisClient.HMGet(table.GetTableKey(), fields...).Result()
	if err != nil {
		return 0, err
	} else if valAry == nil {
		return 0, nil
	}
	if len(fields) != len(valAry) {
		return 0, Err_FieldValueInvalid
	}
	//e.Printf("sliceElementType:%v", sliceElementType)
	elemType := sliceElementType
	var isPointer bool
	if elemType.Kind() == reflect.Ptr {
		isPointer = true
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Ptr {
		return 0, Err_NotSupportPointer2Pointer
	}

	//2、
	//newElemFunc := func(fields []string) reflect.Value {
	//	switch elemType.Kind() {
	//	case reflect.Slice:
	//		slice := reflect.MakeSlice(elemType, len(fields), len(fields))
	//		x := reflect.New(slice.Type())
	//		x.Elem().Set(slice)
	//		return x
	//	default:
	//		return reflect.New(elemType)
	//	}
	//}

	for i := 0; i < len(fields); i += len(table.ColumnsSeq) {
		var beanValue reflect.Value
		//1、
		if isPointer {
			beanValue = reflect.New(sliceElementType.Elem())
		} else {
			beanValue = reflect.New(sliceElementType)
		}
		//2、beanValue=newElemFunc(table.ColumnsSeq)
		reflectElemVal := reflect.Indirect(beanValue)
		for j, colName := range table.ColumnsSeq {
			if valAry[i+j] == nil && colName == table.PrimaryKey {
				break
			}
			if valAry[i+j] == nil {
				continue
			}
			colValue := reflectElemVal.FieldByName(colName)
			SetValue(valAry[i+j], &colValue)
		}

		if isPointer {
			if sliceValue.CanSet() {
				sliceValue.Set(reflect.Append(sliceValue, (&beanValue).Elem().Addr()))
			}
		} else {
			if sliceValue.CanSet() {
				sliceValue.Set(reflect.Append(sliceValue, (&beanValue).Elem()))
			}
		}
	}
	return count, nil
}

func (e *Engine) InsertMulti(beans ...interface{}) (int, error) {
	var (
		table *Table
		err   error
	)

	valMap := make(map[string]interface{})

	var affectBeans []interface{}
	for _, bean := range beans {
		beanValue := reflect.ValueOf(bean)
		if beanValue.Kind() != reflect.Ptr {
			return 0, Err_NeedPointer
		}
		reflectVal := reflect.Indirect(beanValue)
		if table == nil {
			var has bool
			table, has = e.GetTableByName(e.TableName(reflectVal))
			if !has {
				e.Printfln("GetTable(%v,%v),!has", beanValue, reflectVal)
				continue
			}
		}
		for colName, col := range table.ColumnsMap {
			colValue := reflectVal.FieldByName(colName)
			if col.IsAutoIncrement || col.IsCombinedIndex || col.IsCreated || col.IsUpdated {
			} else {
				if colValue.IsValid() {
					if ToString(colValue.Interface()) == "" {
						switch colValue.Kind() {
						case reflect.String:
							SetDefaultValue(col, &colValue)
						}
					}
					if ToString(colValue.Interface()) == "0" {
						switch colValue.Kind() {
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
							SetDefaultValue(col, &colValue)
						}
					}
				}
			}
		}
		pkOldId, err := e.Index.IsExistData(table, beanValue, reflectVal)
		if err != nil {
			e.Printfln("IsExistData(%v) err:%v", bean, err)
			continue
		}
		if pkOldId > 0 {
			continue
		}
		var lastId int64
		if table.AutoIncrement != "" {
			lastId, err = e.redisClient.HIncrBy(table.GetTableKey(), table.GetAutoIncrKey(), 1).Result()
			if err != nil {
				e.Printfln("HIncrBy(%v,%v) err:%v", table.GetTableKey(), table.GetAutoIncrKey(), err)
				continue
			}
		}
		for colName, col := range table.ColumnsMap {
			fieldName := GetFieldName(lastId, colName)
			colValue := reflectVal.FieldByName(colName)
			if col.IsAutoIncrement {
				valMap[fieldName] = ToString(lastId)
				if colValue.CanSet() {
					colValue.SetInt(lastId)
				}
			} else if col.IsCombinedIndex {

			} else if col.IsCreated || col.IsUpdated {
				createdAt := time.Now().In(e.TZLocation).Unix()
				valMap[fieldName] = createdAt
				if colValue.CanSet() {
					colValue.SetInt(createdAt)
				}
			} else {
				SetDefaultValue(col, &colValue)
				valMap[fieldName] = ToString(colValue.Interface())
			}
		}
		affectBeans = append(affectBeans, bean)
	}
	if table == nil {
		return 0, ERR_UnKnowTable
	}
	if len(affectBeans) == 0 {
		return 0, Err_DataHadAvailable
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		for _, bean := range affectBeans {
			beanValue := reflect.ValueOf(bean)
			if e.isSync2DB && table.IsSync2DB {
				e.syncDB.Add(bean, db_lazy.LazyOperateType_Insert, nil, "")
			}
			reflectVal := reflect.Indirect(beanValue)
			err = e.Index.Update(table, beanValue, reflectVal)
			if err != nil {
				e.Printfln("InsertMulti Update(%s,%v) err:%v", table.Name, bean, err)
			}
		}
	}
	return len(affectBeans), err
}

//Done:unique index is exist? -> IsExistData
func (e *Engine) Insert(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	if beanValue.Kind() != reflect.Ptr {
		return Err_NeedPointer
	}
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return ERR_UnKnowTable
	}

	pkOldId, err := e.Index.IsExistData(table, beanValue, reflectVal)
	if err != nil {
		return err
	}
	if pkOldId > 0 {
		return Err_DataHadAvailable
	}

	for colName, col := range table.ColumnsMap {
		colValue := reflectVal.FieldByName(colName)
		if col.IsAutoIncrement || col.IsCombinedIndex || col.IsCreated || col.IsUpdated {
		} else {
			if colValue.IsValid() {
				if ToString(colValue.Interface()) == "" {
					switch colValue.Kind() {
					case reflect.String:
						SetDefaultValue(col, &colValue)
					}
				}
				if ToString(colValue.Interface()) == "0" {
					switch colValue.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
						SetDefaultValue(col, &colValue)
					}
				}
			}
		}
	}

	var lastId int64
	if table.AutoIncrement != "" {
		lastId, err = e.redisClient.HIncrBy(table.GetTableKey(), table.GetAutoIncrKey(), 1).Result()
		if err != nil {
			return err
		}
	}else{
		colValue := reflectVal.FieldByName(table.PrimaryKey)
		lastId=colValue.Int()
	}

	valMap := make(map[string]interface{})

	for colName, col := range table.ColumnsMap {
		fieldName := GetFieldName(lastId, colName)
		colValue := reflectVal.FieldByName(colName)
		if col.IsAutoIncrement {
			valMap[fieldName] = ToString(lastId)
			if colValue.CanSet() {
				colValue.SetInt(lastId)
			}
		} else if col.IsCombinedIndex {

		} else if col.IsCreated || col.IsUpdated {
			createdAt := time.Now().In(e.TZLocation).Unix()
			valMap[fieldName] = createdAt
			if colValue.CanSet() {
				colValue.SetInt(createdAt)
			}
		} else {
			SetDefaultValue(col, &colValue)
			valMap[fieldName] = ToString(colValue.Interface())
		}
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		err = e.Index.Update(table, beanValue, reflectVal)
		if err != nil {
			e.Printfln("Insert Update(%s,%v) err:%v", table.Name, bean, err)
		}
		if e.isSync2DB && table.IsSync2DB {
			e.syncDB.Add(bean, db_lazy.LazyOperateType_Insert, nil, "")
		}
	}
	return err
}

func (e *Engine) GetDefaultValue(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return ERR_UnKnowTable
	}

	for colName, col := range table.ColumnsMap {
		colValue := reflectVal.FieldByName(colName)
		SetDefaultValue(col, &colValue)
	}
	return nil
}

func (e *Engine) UpdateMulti(bean interface{}, searchCon *SearchCondition, cols ...string) (int, error) {
	beanValue := reflect.ValueOf(bean)
	if beanValue.Kind() != reflect.Ptr {
		return 0, Err_NeedPointer
	}
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return 0, ERR_UnKnowTable
	}

	cols = e.FileterCols(table, cols...)

	if len(cols) == 0 {
		return 0, ERR_UnKnowField
	}

	idAry, err := e.Index.Range(table, searchCon, 0, 10000)
	if err != nil {
		return 0, err
	}
	if len(idAry) == 0 {
		return 0, nil
	}
	if len(idAry) > 1 {
		for _, col := range cols {
			if index, ok := table.IndexesMap[table.GetIndexKey(col)]; ok {
				if index.Type == IndexType_IdScore { //unique index, can't update more than one data
					return 0, Err_DataHadAvailable
				}
			}
		}

		for _, v := range table.IndexesMap {
			if len(v.IndexColumn) > 1 {
				var meetCount int
				for _, colIndex := range v.IndexColumn {
					for _, col := range cols {
						if col == colIndex {
							meetCount++
							break
						}
					}
				}
				if meetCount == len(v.IndexColumn) { //combinedindex, can't update more than one data
					return 0, Err_DataHadAvailable
				}
			}
		}
	}

	valMap := make(map[string]interface{})

	pkOldId, err := e.Index.IsExistData(table, beanValue, reflectVal, cols...)
	if err != nil {
		return 0, err
	}
	e.Printfln("idAry:%v,pkOldId:%d", idAry, pkOldId)
	if pkOldId > 0 {
		if len(idAry) > 1 {
			return 0, Err_DataHadAvailable
		}
		if pkOldId > 0 {
			var pkInt int64
			SetInt64FromStr(&pkInt, idAry[0])
			if pkInt != pkOldId {
				return 0, Err_DataHadAvailable
			}
		}
	}

	for _, pkIntStr := range idAry {
		for _, colUpdate := range cols {
			col, ok := table.ColumnsMap[colUpdate]
			if !ok {
				continue
			}

			if col.IsCreated {
				continue
			} else if col.IsCombinedIndex {
				continue
			} else if col.IsPrimaryKey {
				continue
			}
			if len(idAry) > 0 && col.IsPrimaryKey {

			}

			fieldName := GetFieldName(pkIntStr, colUpdate)
			if col.IsUpdated {
				valMap[fieldName] = time.Now().In(e.TZLocation).Unix()
				continue
			}

			colValue := reflectVal.FieldByName(colUpdate)
			if colValue.IsValid() {
				valMap[fieldName] = ToString(colValue.Interface())
			} else {
				valMap[fieldName] = col.DefaultValue
			}
		}
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		for _, pkIntStr := range idAry {
			var pkInt int64
			SetInt64FromStr(&pkInt, pkIntStr)

			colValue := reflectVal.FieldByName(table.PrimaryKey)
			if colValue.CanSet() {
				colValue.SetInt(pkInt)
			}

			err = e.Index.Update(table, beanValue, reflectVal, cols...)
			if err != nil {
				e.Printfln("UpdateMulti Update(%s) err:%v", table.Name, err)
			}
			if e.isSync2DB && table.IsSync2DB {
				var colsDb []string
				for _, col := range cols {
					colsDb = append(colsDb, Camel2Underline(col))
				}
				e.syncDB.Add(bean, db_lazy.LazyOperateType_Update, colsDb, fmt.Sprintf("id=%d", pkInt))
			}
		}
	}

	return len(idAry), nil
}
func (e *Engine) Incr(bean interface{}, col string, val int64) (int64, error) {
	beanValue := reflect.ValueOf(bean)
	if beanValue.Kind() != reflect.Ptr {
		return 0, Err_NeedPointer
	}
	reflectVal := reflect.Indirect(beanValue)

	e.Printfln("incr:%v,%v",beanValue,reflectVal)
	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return 0, ERR_UnKnowTable
	}

	if col == "" {
		return 0, ERR_UnKnowField
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return 0, Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()
	has, err := e.Index.IdIsExist(table, pkInt)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, Err_DataNotAvailable
	}

	//pkOldId, err := e.Index.IsExistData(table, beanValue, reflectVal, table.PrimaryKey)
	//if err != nil {
	//	return 0, err
	//}
	//if pkOldId == 0 {
	//	return 0, Err_DataNotAvailable
	//}

	res, err := e.redisClient.HIncrBy(table.GetTableKey(), GetFieldName(pkInt, col), val).Result()
	if err == nil {
		if e.isSync2DB && table.IsSync2DB {
			colValue := reflectVal.FieldByName(col)
			if colValue.CanSet() {
				colValue.SetInt(res)
			}
			e.syncDB.Add(bean, db_lazy.LazyOperateType_Update, []string{Camel2Underline(col)}, fmt.Sprintf("id=%d", pkInt))
		}
	}
	return res, err
}
func (e *Engine) Sum(bean interface{}, searchCon *SearchCondition, col string) (int64, error) {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return 0, ERR_UnKnowTable
	}

	idAry, err := e.Index.Range(table, searchCon, 0, 10000)
	if err != nil {
		return 0, err
	}
	if len(idAry) == 0 {
		return 0, nil
	}
	fields := make([]string, 0)
	for _, pkIntStr := range idAry {
		fieldName := GetFieldName(pkIntStr, col)
		fields = append(fields, fieldName)
	}
	valAry, err := e.redisClient.HMGet(table.GetTableKey(), fields...).Result()
	if err != nil {
		return 0, err
	} else if valAry == nil {
		return 0, nil
	}
	if len(fields) != len(valAry) {
		return 0, Err_FieldValueInvalid
	}
	var res int64
	for _, val := range valAry {
		var valInt int64
		SetInt64FromStr(&valInt, ToString(val))
		res += valInt
	}
	return res, nil
}
func (e *Engine) Update(bean interface{}, cols ...string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return ERR_UnKnowTable
	}

	cols = e.FileterCols(table, cols...)

	if len(cols) == 0 {
		return ERR_UnKnowField
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	pkOldId, err := e.Index.IsExistData(table, beanValue, reflectVal, cols...)
	if err != nil {
		return err
	}
	if pkOldId > 0 && pkOldId != pkInt {
		return Err_DataHadAvailable
	}

	pkOldId, err = e.Index.IsExistData(table, beanValue, reflectVal, table.PrimaryKey)
	if err != nil {
		return err
	}
	if pkOldId == 0 {
		return Err_DataNotAvailable
	}
	if pkOldId != pkInt {
		return Err_DataHadAvailable
	}

	valMap := make(map[string]interface{})

	for _, colUpdate := range cols {
		col, ok := table.ColumnsMap[colUpdate]
		if !ok {
			continue
		}
		if col.IsCreated {
			continue
		} else if col.IsCombinedIndex {
			continue
		} else if col.IsPrimaryKey {
			continue
		}

		fieldName := GetFieldName(pkInt, colUpdate)

		if col.IsUpdated {
			valMap[fieldName] = time.Now().In(e.TZLocation).Unix()
			continue
		}

		colValue := reflectVal.FieldByName(colUpdate)
		if colValue.IsValid() {
			valMap[fieldName] = ToString(colValue.Interface())
		} else {
			valMap[fieldName] = col.DefaultValue
		}
	}
	if len(valMap) == 0 {
		return nil
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		err = e.Index.Update(table, beanValue, reflectVal, cols...)
		if err != nil {
			e.Printfln("Update Update(%s) err:%v", table.Name, err)
		}
		if e.isSync2DB && table.IsSync2DB {
			var colsDb []string
			for _, col := range cols {
				colsDb = append(colsDb, Camel2Underline(col))
			}
			e.syncDB.Add(bean, db_lazy.LazyOperateType_Update, colsDb, fmt.Sprintf("id=%d", pkOldId))
		}
	}
	return err
}
func (e *Engine) DeleteByCondition(bean interface{}, searchCon *SearchCondition) (int, error) {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return 0, ERR_UnKnowTable
	}

	idAry, err := e.Index.Range(table, searchCon, 0, 10000)
	if err != nil {
		return 0, err
	}
	if len(idAry) == 0 {
		return 0, nil
	}

	fields := make([]string, 0)

	for _, idStr := range idAry {
		var pkInt int64
		SetInt64FromStr(&pkInt, idStr)
		for _, colName := range table.ColumnsSeq {
			fieldName := GetFieldName(pkInt, colName)
			fields = append(fields, fieldName)
		}
	}

	_, err = e.redisClient.HDel(table.GetTableKey(), fields...).Result()
	if err == nil {
		for _, idStr := range idAry {
			var pkInt int64
			SetInt64FromStr(&pkInt, idStr)
			e.Index.Delete(table, pkInt)
			if e.isSync2DB && table.IsSync2DB {
				e.syncDB.Add(bean, db_lazy.LazyOperateType_Delete, nil, fmt.Sprintf("id=%d", pkInt))
			}
		}
	}
	return len(idAry), nil
}
func (e *Engine) Delete(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return ERR_UnKnowTable
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()
	has, err := e.Index.IdIsExist(table, pkInt)
	if err != nil {
		return err
	}
	if !has {
		return Err_DataNotAvailable
	}
	//getId, err := e.Index.GetId(table, &SearchCondition{
	//	SearchColumn: []string{table.PrimaryKey},
	//	//IndexType:     IndexType_IdMember,
	//	FieldMinValue: pkInt,
	//	FieldMaxValue: pkInt,
	//})
	//if err != nil {
	//	return err
	//}
	//if getId == 0 {
	//	return Err_DataNotAvailable
	//}

	fields := make([]string, 0)
	for _, colName := range table.ColumnsSeq {
		fieldName := GetFieldName(pkInt, colName)
		fields = append(fields, fieldName)
	}

	_, err = e.redisClient.HDel(table.GetTableKey(), fields...).Result()
	if err == nil {
		e.Index.Delete(table, pkInt)
		if e.isSync2DB && table.IsSync2DB {
			e.syncDB.Add(bean, db_lazy.LazyOperateType_Delete, nil, fmt.Sprintf("id=%d", pkInt))
		}
	}
	return nil
}

//del the hashkey, it will del all elements for this hash
func (e *Engine) TableTruncate(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, has := e.GetTableByName(e.TableName(reflectVal))
	if !has {
		return ERR_UnKnowTable
	}
	_, err := e.redisClient.Del(table.GetTableKey()).Result()
	if err != nil {
		return err
	}
	err = e.Index.Drop(table)
	//todo:TableTruncate syncDB
	return err
}
