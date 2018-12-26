package redis_orm

import (
	"reflect"
	"time"
)

/*
only support one searchCondition to get or find
todo: SearchCondition not a elegant way..
*/
func (e *Engine) GetByCondition(bean interface{}, searchCon *SearchCondition) (bool, error) {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return false, err
	}

	getId, err := e.indexGetId(table, searchCon)
	if err != nil {
		return false, err
	}
	if getId == 0 {
		return false, Err_DataNotAvailable
	}
	colValue := reflectVal.FieldByName(table.PrimaryKey)
	colValue.SetInt(getId)

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
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return false, err
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return false, Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	getId, err := e.indexGetId(table, &SearchCondition{
		SearchColumn:  []string{table.PrimaryKey},
		IndexType:     IndexType_IdMember,
		FieldMinValue: pkInt,
		FieldMaxValue: pkInt,
	})
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
		err        error
		reflectVal reflect.Value
	)
	sliceElementType := sliceValue.Type().Elem()
	if sliceElementType.Kind() == reflect.Ptr {
		if sliceElementType.Elem().Kind() == reflect.Struct {
			beanValue := reflect.New(sliceElementType.Elem())
			reflectVal = reflect.Indirect(beanValue)
			table, err = e.GetTable(beanValue, reflectVal)
			if err != nil {
				return 0, err
			}
		}
	} else if sliceElementType.Kind() == reflect.Struct {
		beanValue := reflect.New(sliceElementType)
		reflectVal = reflect.Indirect(beanValue)
		table, err = e.GetTable(beanValue, reflectVal)
		if err != nil {
			return 0, err
		}
	}
	if table == nil {
		return 0, Err_UnSupportedTableModel
	}
	return e.count(searchCon, table)
}
func (e *Engine) count(searchCon *SearchCondition, table *Table) (int64, error) {
	count, err := e.indexCount(table, searchCon)
	return count, err
}
func (e *Engine) Find(offset, limit int64, searchCon *SearchCondition, beanAry interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(beanAry))
	if sliceValue.Kind() != reflect.Slice {
		return 0, Err_NeedPointer
	}

	var (
		table      *Table
		err        error
		reflectVal reflect.Value
	)
	sliceElementType := sliceValue.Type().Elem()
	if sliceElementType.Kind() == reflect.Ptr {
		if sliceElementType.Elem().Kind() == reflect.Struct {
			beanValue := reflect.New(sliceElementType.Elem())
			reflectVal = reflect.Indirect(beanValue)
			table, err = e.GetTable(beanValue, reflectVal)
			if err != nil {
				return 0, err
			}
		}
	} else if sliceElementType.Kind() == reflect.Struct {
		beanValue := reflect.New(sliceElementType)
		reflectVal = reflect.Indirect(beanValue)
		table, err = e.GetTable(beanValue, reflectVal)
		if err != nil {
			return 0, err
		}
	}
	if table == nil {
		return 0, Err_UnSupportedTableModel
	}
	count, err := e.count(searchCon, table)
	if err != nil {
		return 0, nil
	}
	idAry, err := e.indexRange(table, searchCon, offset, limit)
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

//Done:unique index is exist? -> indexIsExistData
func (e *Engine) Insert(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return err
	}
	pkOldId, err := e.indexIsExistData(table, beanValue, reflectVal)
	if err != nil {
		return err
	}
	if pkOldId > 0 {
		return Err_DataHadAvailable
	}
	var lastId int64
	if table.AutoIncrement != "" {
		lastId, err = e.redisClient.HIncrBy(table.GetTableKey(), table.GetAutoIncrKey(), 1).Result()
		if err != nil {
			return err
		}
	}

	valMap := make(map[string]interface{})

	for colName, col := range table.ColumnsMap {
		fieldName := GetFieldName(lastId, colName)
		colValue := reflectVal.FieldByName(colName)
		if col.IsAutoIncrement {
			valMap[fieldName] = ToString(lastId)
			colValue.SetInt(lastId)
		} else if col.IsCombinedIndex {

		} else if col.IsCreated || col.IsUpdated {
			createdAt := time.Now().In(e.TZLocation).Unix()
			valMap[fieldName] = createdAt
			colValue.SetInt(createdAt)
		} else {
			SetDefaultValue(col, &colValue)
			valMap[fieldName] = ToString(colValue.Interface())
		}
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		err = e.indexUpdate(table, beanValue, reflectVal)
		if err != nil {
			e.Printfln("Insert indexUpdate(%s) err:%v", table.Name, err)
		}
	}
	return err
}

func (e *Engine) GetDefaultValue(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return err
	}
	for colName, col := range table.ColumnsMap {
		colValue := reflectVal.FieldByName(colName)
		SetDefaultValue(col, &colValue)
	}
	return nil
}

func (e *Engine) UpdateMulti(bean interface{}, searchCon *SearchCondition, cols ...string) (int, error) {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return 0, err
	}
	if len(cols) == 0 {
		cols = table.ColumnsSeq
	}

	idAry, err := e.indexRange(table, searchCon, 0, 10000)
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

	pkOldId, err := e.indexIsExistData(table, beanValue, reflectVal, cols...)
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
			colValue.SetInt(pkInt)

			err = e.indexUpdate(table, beanValue, reflectVal, cols...)
			if err != nil {
				e.Printfln("UpdateMulti indexUpdate(%s) err:%v", table.Name, err)
			}
		}
	}

	return len(idAry), nil
}
func (e *Engine) Incr(bean interface{}, col string, val int64) (int64, error) {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return 0, err
	}
	if col == "" {
		return 0, ERR_UnKnowField
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return 0, Err_PrimaryKeyTypeInvalid
	}

	pkOldId, err := e.indexIsExistData(table, beanValue, reflectVal, table.PrimaryKey)
	if err != nil {
		return 0, err
	}
	if pkOldId == 0 {
		return 0, Err_DataNotAvailable
	}

	res, err := e.redisClient.HIncrBy(table.GetTableKey(), GetFieldName(pkOldId, col), val).Result()
	return res, err
}
func (e *Engine) Update(bean interface{}, cols ...string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		cols = table.ColumnsSeq
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	pkOldId, err := e.indexIsExistData(table, beanValue, reflectVal, cols...)
	if err != nil {
		return err
	}
	if pkOldId > 0 && pkOldId != pkInt {
		return Err_DataHadAvailable
	}

	pkOldId, err = e.indexIsExistData(table, beanValue, reflectVal, table.PrimaryKey)
	if err != nil {
		return err
	}
	if pkOldId == 0 {
		return Err_DataNotAvailable
	}
	if pkOldId != pkInt {
		return Err_DataHadAvailable
	}

	//getId, err := e.indexGetId(&SearchCondition{
	//	SearchColumn:  []string{table.PrimaryKey},
	//	IndexType:     IndexType_IdMember,
	//	FieldMinValue: pkInt,
	//	FieldMaxValue: pkInt,
	//}, bean)
	//if err != nil {
	//	return err
	//}
	//if getId == 0 {
	//	return Err_DataNotAvailable
	//}

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
		err = e.indexUpdate(table, beanValue, reflectVal, cols...)
		if err != nil {
			e.Printfln("Update indexUpdate(%s) err:%v", table.Name, err)
		}
	}
	return err
}
func (e *Engine) DeleteMulti(bean interface{}, searchCon *SearchCondition, cols ...string) (int, error) {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return 0, err
	}
	if len(cols) == 0 {
		cols = table.ColumnsSeq
	}

	idAry, err := e.indexRange(table, searchCon, 0, 10000)
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
			e.indexDelete(table, pkInt)
		}
	}
	return len(idAry), nil
}
func (e *Engine) Delete(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return err
	}

	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkInt := pkFieldValue.Int()

	getId, err := e.indexGetId(table, &SearchCondition{
		SearchColumn:  []string{table.PrimaryKey},
		IndexType:     IndexType_IdMember,
		FieldMinValue: pkInt,
		FieldMaxValue: pkInt,
	})
	if err != nil {
		return err
	}
	if getId == 0 {
		return Err_DataNotAvailable
	}

	fields := make([]string, 0)
	for _, colName := range table.ColumnsSeq {
		fieldName := GetFieldName(pkInt, colName)
		fields = append(fields, fieldName)
	}

	_, err = e.redisClient.HDel(table.GetTableKey(), fields...).Result()
	if err == nil {
		e.indexDelete(table, pkInt)
	}
	return nil
}

//del the hashkey, it will del all elements for this hash
func (e *Engine) TableDrop(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return err
	}
	_, err = e.redisClient.Del(table.GetTableKey()).Result()
	if err != nil {
		return err
	}
	err = e.indexDrop(table)
	return err
}

//del the hashkey, it will del all elements for this hash
func (e *Engine) TableTruncate(bean interface{}) error {
	return e.TableDrop(bean)
}
