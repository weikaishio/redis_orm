package redis_orm

import (
	"time"
)

func (e *Engine) DeleteByPK(table *Table, pkInt int64) error {
	fields := make([]string, 0)

	for _, colName := range table.ColumnsSeq {
		fieldName := GetFieldName(pkInt, colName)
		fields = append(fields, fieldName)
	}

	_, err := e.redisClient.HDel(table.GetTableKey(), fields...).Result()
	if err == nil {
		e.Index.Delete(table, pkInt)
		//todo:support syncDB.Add, need general SQL
		//if e.isSync2DB && table.IsSync2DB {
		//	e.syncDB.Add(bean, db_lazy.LazyOperateType_Delete, nil, fmt.Sprintf("id=%d", pkInt))
		//}
	}
	return err
}

func (e *Engine) UpdateByMap(table *Table, columnValMap map[string]string) error {
	var cols []string
	for col, _ := range columnValMap {
		_, isExist := table.ColumnsMap[col]
		if !isExist {
			cols = append(cols, col)
		}
	}

	if len(cols) == 0 {
		return ERR_UnKnowField
	}

	pkFieldValue, ok := columnValMap[table.PrimaryKey]
	if !ok {
		return Err_PrimaryKeyNotFound
	}
	var pkInt int64
	SetInt64FromStr(&pkInt, pkFieldValue)
	if pkInt <= 0 {
		return Err_PrimaryKeyTypeInvalid
	}

	pkOldId, err := e.Index.IsExistDataByMap(table, columnValMap)
	if err != nil {
		return err
	}
	if pkOldId > 0 && pkOldId != pkInt {
		return Err_DataHadAvailable
	}
	if pkOldId == 0 {
		return Err_DataNotAvailable
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

		if val, ok := columnValMap[colUpdate]; ok {
			valMap[fieldName] = val
		} else {
			valMap[fieldName] = col.DefaultValue
		}
	}
	if len(valMap) == 0 {
		return nil
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		err = e.Index.UpdateByMap(table, pkInt, columnValMap)
		if err != nil {
			e.Printfln("Update Update(%s) err:%v", table.Name, err)
		}
		//todo:support syncDB.Add
		//if e.isSync2DB && table.IsSync2DB {
		//	var colsDb []string
		//	for _, col := range cols {
		//		colsDb = append(colsDb, Camel2Underline(col))
		//	}
		//	e.syncDB.Add(bean, db_lazy.LazyOperateType_Update, colsDb, fmt.Sprintf("id=%d", pkOldId))
		//}
	}
	return err
}

func (e *Engine) InsertByMap(table *Table, columnValMap map[string]string) error {
	var cols []string
	for col, _ := range columnValMap {
		_, isExist := table.ColumnsMap[col]
		if !isExist {
			cols = append(cols, col)
		}
	}

	if len(cols) == 0 {
		return ERR_UnKnowField
	}

	pkOldId, err := e.Index.IsExistDataByMap(table, columnValMap)
	if err != nil {
		return err
	}
	if pkOldId > 0 {
		return Err_DataHadAvailable
	}

	for colName, col := range table.ColumnsMap {
		if col.IsAutoIncrement || col.IsCombinedIndex || col.IsCreated || col.IsUpdated {
		} else {
			_, ok := columnValMap[colName]
			if !ok {
				columnValMap[colName] = col.DefaultValue
			}
		}
	}

	var lastId int64
	if table.AutoIncrement != "" {
		lastId, err = e.redisClient.HIncrBy(table.GetTableKey(), table.GetAutoIncrKey(), 1).Result()
		if err != nil {
			return err
		}
	} else {
		pkFieldValue, ok := columnValMap[table.PrimaryKey]
		if !ok {
			return Err_PrimaryKeyNotFound
		}
		SetInt64FromStr(&lastId, pkFieldValue)
	}

	valMap := make(map[string]interface{})

	for colName, col := range table.ColumnsMap {
		fieldName := GetFieldName(lastId, colName)
		colValue, _ := columnValMap[colName]
		if col.IsAutoIncrement {
			valMap[fieldName] = ToString(lastId)
			//if colValue.CanSet() {
			//	colValue.SetInt(lastId)
			//}
		} else if col.IsCombinedIndex {

		} else if col.IsCreated || col.IsUpdated {
			createdAt := time.Now().In(e.TZLocation).Unix()
			valMap[fieldName] = createdAt
			//if colValue.CanSet() {
			//	colValue.SetInt(createdAt)
			//}
		} else {
			//SetDefaultValue(col, &colValue)
			valMap[fieldName] = colValue
		}
	}
	_, err = e.redisClient.HMSet(table.GetTableKey(), valMap).Result()
	if err == nil {
		err = e.Index.UpdateByMap(table, lastId, columnValMap)
		if err != nil {
			e.Printfln("Insert Update(%s,%v) err:%v", table.Name, columnValMap, err)
		}
		//todo:support syncDB.Add
		//if e.isSync2DB && table.IsSync2DB {
		//	e.syncDB.Add(bean, db_lazy.LazyOperateType_Insert, nil, "")
		//}
	}
	return err
}
