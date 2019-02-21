package redis_orm

import (
	"fmt"
	"reflect"

	"github.com/go-redis/redis"
	"strings"
)

type IndexEngine struct {
	engine *Engine
}

func NewIndexEngine(e *Engine) *IndexEngine {
	return &IndexEngine{
		engine: e,
	}
}

//Done:combined index
func (ixe *IndexEngine) GetId(table *Table, searchCon *SearchCondition) (int64, error) {
	index, ok := table.IndexesMap[strings.ToLower(searchCon.Name())]
	if !ok {
		err := fmt.Errorf("searchCon.index:%s not exist\n", strings.ToLower(searchCon.Name()))
		return 0, err
	}
	switch index.Type {
	case IndexType_IdMember:
		redisZ := redis.ZRangeBy{
			Min:    ToString(searchCon.FieldMinValue),
			Max:    ToString(searchCon.FieldMaxValue),
			Offset: 0,
			Count:  1,
		}
		res, err := ixe.engine.redisClient.ZRangeByScore(index.NameKey, redisZ).Result()
		if err == nil {
			if len(res) == 0 {
				return 0, nil
			}
			var id int64
			SetInt64FromStr(&id, res[0])
			return id, nil
		} else {
			ixe.engine.Printfln("GetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
			return 0, err
		}
	case IndexType_IdScore:
		res, err := ixe.engine.redisClient.ZScore(index.NameKey, ToString(searchCon.FieldMinValue)).Result()
		if err == nil || (err != nil && err.Error() == redis.Nil.Error()) {
			return int64(res), nil
		} else {
			ixe.engine.Printfln("GetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
			return 0, err
		}
	case IndexType_UnSupport:
		ixe.engine.Printfln("GetId unsupport index type")
	}
	return 0, nil
}
func (ixe *IndexEngine) Count(table *Table, searchCon *SearchCondition) (count int64, err error) {
	index, ok := table.IndexesMap[strings.ToLower(searchCon.Name())]
	if !ok {
		err = fmt.Errorf("searchCon.index:%s not exist\n", strings.ToLower(searchCon.Name()))
		return
	}
	switch index.Type {
	case IndexType_IdMember:
		count, err = ixe.engine.redisClient.ZCount(index.NameKey, ToString(searchCon.FieldMinValue), ToString(searchCon.FieldMaxValue)).Result()
		if err != nil {
			ixe.engine.Printfln("GetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
		}
		return
	case IndexType_IdScore:
		var res float64
		res, err = ixe.engine.redisClient.ZScore(index.NameKey, ToString(searchCon.FieldMinValue)).Result()
		if err == nil || (err != nil && err.Error() == redis.Nil.Error()) {
			if err == nil && res > 0 {
				count = 1
			}
		} else {
			ixe.engine.Printfln("GetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
		}
		return
	case IndexType_UnSupport:
		ixe.engine.Printfln("GetId unsupport index type")
	}
	return
}
func (ixe *IndexEngine) Range(table *Table, searchCon *SearchCondition, offset, count int64) (idAry []string, err error) {
	index, ok := table.IndexesMap[strings.ToLower(searchCon.Name())]
	if !ok {
		err = fmt.Errorf("searchCon.index:%s not exist\n", strings.ToLower(searchCon.Name()))
		return
	}
	switch index.Type {
	case IndexType_IdMember:
		redisZ := redis.ZRangeBy{
			Min:    ToString(searchCon.FieldMinValue),
			Max:    ToString(searchCon.FieldMaxValue),
			Offset: offset,
			Count:  count,
		}
		if searchCon.IsAsc {
			idAry, err = ixe.engine.redisClient.ZRangeByScore(index.NameKey, redisZ).Result()
		} else {
			idAry, err = ixe.engine.redisClient.ZRevRangeByScore(index.NameKey, redisZ).Result()
		}
		if err != nil {
			ixe.engine.Printfln("GetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
		}

		return idAry, nil
	case IndexType_IdScore:
		var res float64
		res, err = ixe.engine.redisClient.ZScore(index.NameKey, ToString(searchCon.FieldMinValue)).Result()
		if err == nil || (err != nil && err.Error() == redis.Nil.Error()) {
			idAry = append(idAry, ToString(res))
		} else {
			ixe.engine.Printfln("GetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
		}
		return idAry, nil
	case IndexType_UnSupport:
		ixe.engine.Printfln("GetId unsupport index type")
	}
	return idAry, nil
}
func (ixe *IndexEngine) Delete(table *Table, idInt int64) error {
	if table.PrimaryKey == "" {
		return Err_PrimaryKeyNotFound
	}
	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		switch index.Type {
		case IndexType_IdMember:
			_, err := ixe.engine.redisClient.ZRem(index.NameKey, idInt).Result()
			if err != nil {
				ixe.engine.Printfln("Delete %s:%v,err:%v", index.NameKey, idInt, err)
				return err
			}
		case IndexType_IdScore:
			_, err := ixe.engine.redisClient.ZRemRangeByScores(index.NameKey, ToString(idInt), ToString(idInt)).Result()
			if err != nil {
				ixe.engine.Printfln("Delete ZRemRangeByScores %s:%v,err:%v", index.NameKey, ToString(idInt), err)
				return err
			}
		case IndexType_UnSupport:
			ixe.engine.Printfln("Delete unsupport index type")
		}
	}
	return nil
}
func ColsIsExistIndex(index *Index, cols ...string) bool {
	isExist := false
	if len(index.IndexColumn) == 1 {
		for _, col := range cols {
			if index.IndexColumn[0] == col {
				isExist = true
				break
			}
		}
	} else {
		var meetCount int
		for _, colIndex := range index.IndexColumn {
			for _, col := range cols {
				if col == colIndex {
					meetCount++
					break
				}
			}
		}
		if meetCount > 0 {
			isExist = true
		}
	}
	return isExist
}

func (ixe *IndexEngine) IdIsExist(table *Table, pkId int64) (bool, error) {
	fieldName := GetFieldName(pkId, table.PrimaryKey)
	val, err := ixe.engine.redisClient.HGet(table.GetTableKey(), fieldName).Result()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			return false, nil
		}
		return false, err
	}
	return val != "", nil
}

//当前数据是否已经存在，存在则返回主键ID，唯一索引的才需要判断是否存在！
func (ixe *IndexEngine) IsExistData(table *Table, beanValue, reflectVal reflect.Value, cols ...string) (int64, error) {
	typ := reflectVal.Type()
	_, has := typ.FieldByName(table.PrimaryKey)
	if !has {
		return 0, Err_PrimaryKeyNotFound
	}

	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		if len(cols) > 0 {
			if !ColsIsExistIndex(index, cols...) {
				continue
			}
		}
		var fieldValueAry []reflect.Value
		for _, column := range index.IndexColumn {
			fieldValue := reflectVal.FieldByName(column)
			fieldValueAry = append(fieldValueAry, fieldValue)
		}
		switch index.Type {
		case IndexType_IdMember:
			if !index.IsUnique && index.NameKey != table.GetIndexKey(table.PrimaryKey) {
				ixe.engine.Printfln("!index.IsUnique break")
				break
			}
			if len(index.IndexColumn) > 2 {
				return 0, Err_CombinedIndexColCountOver
			}
			var score int64
			switch fieldValueAry[0].Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				score = fieldValueAry[0].Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				score = int64(fieldValueAry[0].Uint())
			case reflect.Float32, reflect.Float64:
				score = int64(fieldValueAry[0].Float())
			case reflect.String:
				SetInt64FromStr(&score, fieldValueAry[0].String())
			default:
				SetInt64FromStr(&score, fmt.Sprintf("%v", fieldValueAry[0].Interface()))
			}
			if len(fieldValueAry) == 2 {
				score = score << 32
				switch fieldValueAry[1].Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					score = score | fieldValueAry[1].Int()
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					score = score | int64(fieldValueAry[1].Uint())
				case reflect.Float32, reflect.Float64:
					score = score | int64(fieldValueAry[1].Float())
				case reflect.String:
					var subScore int32
					SetInt32FromStr(&subScore, fieldValueAry[1].String())
					score = score | int64(subScore)
				default:
					var subScore int32
					SetInt32FromStr(&subScore, fmt.Sprintf("%v", fieldValueAry[1].Interface()))
					score = score | int64(subScore)
				}
			}
			zRangeBy := redis.ZRangeBy{
				Min:    ToString(score),
				Max:    ToString(score),
				Offset: 0,
				Count:  1,
			}
			val, err := ixe.engine.redisClient.ZRangeByScore(index.NameKey, zRangeBy).Result()
			if err != nil {
				ixe.engine.Printfln("IsExistData ZRangeByScore%s:%v,err:%v", index.NameKey, score, err)
				return 0, err
			} else if len(val) > 0 {
				var pkOldId int64
				SetInt64FromStr(&pkOldId, val[0])
				if pkOldId > 0 {
					return pkOldId, nil
				}
			}
		case IndexType_IdScore:
			var members []string
			for _, fieldValue := range fieldValueAry {
				if fieldValue.IsValid() {
					members = append(members, ToString(fieldValue.Interface()))
				}
			}
			//log.Trace("IndexUpdate %s:%v", index.NameKey, redisZ)
			pkOldId, err := ixe.engine.redisClient.ZScore(index.NameKey, strings.Join(members, "&")).Result()
			if err != nil && err.Error() != redis.Nil.Error() {
				ixe.engine.Printfln("IsExistData %s:%v,err:%v", index.NameKey, strings.Join(members, "&"), err)
				return 0, err
			} else {
				ixe.engine.Printfln("ZScore(%s,%s) pkOldId:%d", index.NameKey, strings.Join(members, "&"), int64(pkOldId))
				if int64(pkOldId) > 0 {
					return int64(pkOldId), nil
				}
			}
		case IndexType_UnSupport:
			ixe.engine.Printfln("IsExistData unsupport index type")
		}
	}
	return 0, nil
}
func (ixe *IndexEngine) IsExistDataByMap(table *Table, valMap map[string]string) (int64, error) {
	var cols []string
	for k, _ := range valMap {
		cols = append(cols, k)
	}

	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		if len(cols) > 0 {
			if !ColsIsExistIndex(index, cols...) {
				continue
			}
		}

		var fieldValueAry []string
		for _, column := range index.IndexColumn {
			if fieldValue, ok := valMap[column]; ok {
				fieldValueAry = append(fieldValueAry, fieldValue)
			}
		}
		switch index.Type {
		case IndexType_IdMember:
			if !index.IsUnique && index.NameKey != table.GetIndexKey(table.PrimaryKey) {
				ixe.engine.Printfln("!index.IsUnique break")
				break
			}
			if len(index.IndexColumn) > 2 {
				return 0, Err_CombinedIndexColCountOver
			}
			var score int64
			SetInt64FromStr(&score, fieldValueAry[0])

			if len(fieldValueAry) == 2 {
				score = score << 32
				var subScore int32
				SetInt32FromStr(&subScore, fieldValueAry[1])
				score = score | int64(subScore)
			}
			zRangeBy := redis.ZRangeBy{
				Min:    ToString(score),
				Max:    ToString(score),
				Offset: 0,
				Count:  1,
			}
			val, err := ixe.engine.redisClient.ZRangeByScore(index.NameKey, zRangeBy).Result()
			if err != nil {
				ixe.engine.Printfln("IsExistData ZRangeByScore%s:%v,err:%v", index.NameKey, score, err)
				return 0, err
			} else if len(val) > 0 {
				var pkOldId int64
				SetInt64FromStr(&pkOldId, val[0])
				if pkOldId > 0 {
					return pkOldId, nil
				}
			}
		case IndexType_IdScore:
			var members []string
			for _, fieldValue := range fieldValueAry {
				members = append(members, fieldValue)

			}
			pkOldId, err := ixe.engine.redisClient.ZScore(index.NameKey, strings.Join(members, "&")).Result()
			if err != nil && err.Error() != redis.Nil.Error() {
				ixe.engine.Printfln("IsExistData %s:%v,err:%v", index.NameKey, strings.Join(members, "&"), err)
				return 0, err
			} else {
				ixe.engine.Printfln("ZScore(%s,%s) pkOldId:%d", index.NameKey, strings.Join(members, "&"), int64(pkOldId))
				if int64(pkOldId) > 0 {
					return int64(pkOldId), nil
				}
			}
		case IndexType_UnSupport:
			ixe.engine.Printfln("IsExistData unsupport index type")
		}
	}
	return 0, nil
}

//todo: no thread safety! watch?
func (ixe *IndexEngine) Update(table *Table, beanValue, reflectVal reflect.Value, cols ...string) error {
	typ := reflectVal.Type()
	_, has := typ.FieldByName(table.PrimaryKey)
	if !has {
		return Err_PrimaryKeyNotFound
	}
	pkFieldValue := reflectVal.FieldByName(table.PrimaryKey)
	if pkFieldValue.Kind() != reflect.Int64 {
		return Err_PrimaryKeyTypeInvalid
	}
	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		if len(cols) > 0 {
			if !ColsIsExistIndex(index, cols...) {
				//ixe.engine.Printfln("Update ColsIsExistIndex:%v,cols:%v", index.IndexColumn, cols)
				continue
			}
		}
		var fieldValueAry []reflect.Value
		for _, column := range index.IndexColumn {
			fieldValue := reflectVal.FieldByName(column)
			fieldValueAry = append(fieldValueAry, fieldValue)
		}
		switch index.Type {
		case IndexType_IdMember:
			if len(index.IndexColumn) > 2 {
				return Err_CombinedIndexColCountOver
			}
			var score int64
			switch fieldValueAry[0].Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				score = fieldValueAry[0].Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				score = int64(fieldValueAry[0].Uint())
			case reflect.Float32, reflect.Float64:
				score = int64(fieldValueAry[0].Float())
			case reflect.String:
				SetInt64FromStr(&score, fieldValueAry[0].String())
			default:
				SetInt64FromStr(&score, fmt.Sprintf("%v", fieldValueAry[0].Interface()))
			}
			if len(fieldValueAry) == 2 {
				score = score << 32
				switch fieldValueAry[1].Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					score = score | fieldValueAry[1].Int()
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					score = score | int64(fieldValueAry[1].Uint())
				case reflect.Float32, reflect.Float64:
					score = score | int64(fieldValueAry[1].Float())
				case reflect.String:
					var subScore int32
					SetInt32FromStr(&subScore, fieldValueAry[1].String())
					score = score | int64(subScore)
				default:
					var subScore int32
					SetInt32FromStr(&subScore, fmt.Sprintf("%v", fieldValueAry[1].Interface()))
					score = score | int64(subScore)
				}
			}
			redisZ := redis.Z{
				Member: pkFieldValue.Int(),
				Score:  float64(score), //todo:浮点数有诡异问题
			}
			_, err := ixe.engine.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				ixe.engine.Printfln("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
				return err
			}
		case IndexType_IdScore:
			//remove old index
			_, err := ixe.engine.redisClient.ZRemRangeByScores(index.NameKey, ToString(pkFieldValue.Int()), ToString(pkFieldValue.Int())).Result()
			if err != nil {
				ixe.engine.Printfln("IndexUpdate ZRemRangeByScores %s:%v,err:%v", index.NameKey, ToString(pkFieldValue.Int()), err)
				return err
			}
			var members []string
			for _, fieldValue := range fieldValueAry {
				if fieldValue.IsValid() {
					members = append(members, ToString(fieldValue.Interface()))
				}
			}
			redisZ := redis.Z{
				Member: strings.Join(members, "&"),
				Score:  float64(pkFieldValue.Int()),
			}
			_, err = ixe.engine.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				ixe.engine.Printfln("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
				return err
			}
		case IndexType_UnSupport:
			ixe.engine.Printfln("IndexUpdate unsupport index type")
		}
	}
	return nil
}

func (ixe *IndexEngine) UpdateByMap(table *Table, pkInt int64, valMap map[string]string) error {
	var cols []string
	for k, _ := range valMap {
		cols = append(cols, k)
	}

	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		if len(cols) > 0 {
			if !ColsIsExistIndex(index, cols...) {
				continue
			}
		}
		var fieldValueAry []string
		for _, column := range index.IndexColumn {
			if fieldValue, ok := valMap[column]; ok {
				fieldValueAry = append(fieldValueAry, fieldValue)
			}
		}
		switch index.Type {
		case IndexType_IdMember:
			if len(index.IndexColumn) > 2 {
				return Err_CombinedIndexColCountOver
			}
			var score int64
			SetInt64FromStr(&score, fieldValueAry[0])

			if len(fieldValueAry) == 2 {
				score = score << 32
				var subScore int32
				SetInt32FromStr(&subScore, fieldValueAry[1])
				score = score | int64(subScore)
			}
			redisZ := redis.Z{
				Member: pkInt,
				Score:  float64(score), //todo:浮点数有诡异问题
			}
			_, err := ixe.engine.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				ixe.engine.Printfln("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
				return err
			}
		case IndexType_IdScore:
			_, err := ixe.engine.redisClient.ZRemRangeByScores(index.NameKey, ToString(pkInt), ToString(pkInt)).Result()
			if err != nil {
				ixe.engine.Printfln("IndexUpdate ZRemRangeByScores %s:%v,err:%v", index.NameKey, ToString(pkInt), err)
				return err
			}
			var members []string
			for _, fieldValue := range fieldValueAry {
				members = append(members, fieldValue)
			}
			redisZ := redis.Z{
				Member: strings.Join(members, "&"),
				Score:  float64(pkInt),
			}
			_, err = ixe.engine.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				ixe.engine.Printfln("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
				return err
			}
		case IndexType_UnSupport:
			ixe.engine.Printfln("IndexUpdate unsupport index type")
		}
	}
	return nil
}

//todo:ReBuild single index
func (ixe *IndexEngine) ReBuild(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)

	table, has := ixe.engine.GetTableByName(ixe.engine.TableName(reflectVal))
	if !has {
		return ERR_UnKnowTable
	}

	ixe.Drop(table, table.PrimaryKey)

	var offset int64 = 0
	var limit int64 = 100

	searchCon := NewSearchCondition(IndexType_IdMember, ScoreMin, ScoreMax, table.PrimaryKey)
	for {
		idAry, err := ixe.Range(table, searchCon, offset, limit)
		if err != nil {
			return err
		}
		if len(idAry) == 0 {
			break
		} else {
			offset += limit
		}

		fields := make([]string, 0)
		for _, id := range idAry {
			for _, colName := range table.ColumnsSeq {
				fieldName := GetFieldName(id, colName)
				fields = append(fields, fieldName)
			}
		}
		valAry, err := ixe.engine.redisClient.HMGet(table.GetTableKey(), fields...).Result()
		if err != nil {
			return err
		} else if valAry == nil {
			break
		}
		if len(fields) != len(valAry) {
			return Err_FieldValueInvalid
		}
		sliceElementType := reflect.TypeOf(bean).Elem()
		for i := 0; i < len(fields); i += len(table.ColumnsSeq) {
			newBeanValue := reflect.New(sliceElementType)
			reflectElemVal := reflect.Indirect(newBeanValue)
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
			err = ixe.Update(table, newBeanValue, reflect.Indirect(newBeanValue))
			if err != nil {
				ixe.engine.Printfln("indexReBuild Update(%v) err:%v", newBeanValue, err)
			}
		}
		if len(idAry) < int(limit) {
			break
		}
	}
	return nil
}
func (ixe *IndexEngine) DropSingleIndex(dropIndex *Index) error {
	_, err := ixe.engine.redisClient.Del(dropIndex.NameKey).Result()
	return err
}
func (ixe *IndexEngine) Drop(table *Table, except ...string) error {
	indexsMap := table.IndexesMap
	var keys []string
	for _, index := range indexsMap {
		isMiss := false
		if len(except) != 0 {
			for _, exceptIndex := range except {
				if exceptIndex == strings.Join(index.IndexColumn, "&") {
					isMiss = true
					break
				}
			}
		}
		if !isMiss {
			keys = append(keys, index.NameKey)
		}
	}
	if len(keys) > 0 {
		_, err := ixe.engine.redisClient.Del(keys...).Result()
		return err
	}
	return nil
}
