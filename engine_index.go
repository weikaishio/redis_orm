package redis_orm

import (
	"fmt"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/mkideal/log"
	"strings"
)

//todo:combined index
func (e *Engine) indexGetId(table *Table, searchCon *SearchCondition) (int64, error) {
	index, ok := table.IndexesMap[strings.ToLower(searchCon.Name())]
	if !ok {
		err := fmt.Errorf("searchCon:%s not exist\n", strings.ToLower(searchCon.Name()))
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
		res, err := e.redisClient.ZRangeByScore(index.NameKey, redisZ).Result()
		if err == nil {
			if len(res) == 0 {
				return 0, nil
			}
			var id int64
			SetInt64FromStr(&id, res[0])
			return id, nil
		} else {
			log.Error("indexGetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
			return 0, err
		}
	case IndexType_IdScore:
		res, err := e.redisClient.ZScore(index.NameKey, ToString(searchCon.FieldMinValue)).Result()
		if err == nil || (err != nil && err.Error() == "redis: nil") {
			return int64(res), nil
		} else {
			log.Error("indexGetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
			return 0, err
		}
	case IndexType_UnSupport:
		log.Error("indexGetId unsupport index type")
	}
	return 0, nil
}
func (e *Engine) indexCount(table *Table, searchCon *SearchCondition) (count int64, err error) {
	index, ok := table.IndexesMap[strings.ToLower(searchCon.Name())]
	if !ok {
		err = fmt.Errorf("searchCon:%s not exist\n", strings.ToLower(searchCon.Name()))
		return
	}
	switch index.Type {
	case IndexType_IdMember:
		count, err = e.redisClient.ZCount(index.NameKey, ToString(searchCon.FieldMinValue), ToString(searchCon.FieldMaxValue)).Result()
		if err != nil {
			log.Error("indexGetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
		}
		return
	case IndexType_IdScore:
		var res float64
		res, err = e.redisClient.ZScore(index.NameKey, ToString(searchCon.FieldMinValue)).Result()
		if err == nil || (err != nil && err.Error() == "redis: nil") {
			if err == nil && res > 0 {
				count = 1
			}
		} else {
			log.Error("indexGetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
		}
		return
	case IndexType_UnSupport:
		log.Error("indexGetId unsupport index type")
	}
	return
}
func (e *Engine) indexRange(table *Table, searchCon *SearchCondition, offset, count int64) (idAry []string, err error) {
	index, ok := table.IndexesMap[strings.ToLower(searchCon.Name())]
	if !ok {
		err = fmt.Errorf("searchCon:%s not exist\n", strings.ToLower(searchCon.Name()))
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
			idAry, err = e.redisClient.ZRangeByScore(index.NameKey, redisZ).Result()
		} else {
			idAry, err = e.redisClient.ZRevRangeByScore(index.NameKey, redisZ).Result()
		}
		if err != nil {
			log.Error("indexGetId ZRangeWithScores(%s,%v,%v) err:%v", index.NameKey, searchCon.FieldMinValue, searchCon.FieldMaxValue, err)
		}

		return idAry, nil
	case IndexType_IdScore:
		var res float64
		res, err = e.redisClient.ZScore(index.NameKey, ToString(searchCon.FieldMinValue)).Result()
		if err == nil || (err != nil && err.Error() == "redis: nil") {
			idAry = append(idAry, ToString(res))
		} else {
			log.Error("indexGetId ZScore(%s,%v) err:%v", index.NameKey, searchCon.FieldMinValue, err)
		}
		return idAry, nil
	case IndexType_UnSupport:
		log.Error("indexGetId unsupport index type")
	}
	return idAry, nil
}
func (e *Engine) indexDelete(table *Table, idInt int64) error {
	if table.PrimaryKey == "" {
		return Err_PrimaryKeyNotFound
	}
	indexsMap := table.IndexesMap
	for _, index := range indexsMap {
		switch index.Type {
		case IndexType_IdMember:
			_, err := e.redisClient.ZRem(index.NameKey, idInt).Result()
			if err != nil {
				log.Warn("indexDelete %s:%v,err:%v", index.NameKey, idInt, err)
				return err
			}
		case IndexType_IdScore:
			_, err := e.redisClient.ZRemRangeByScores(index.NameKey, ToString(idInt), ToString(idInt)).Result()
			if err != nil {
				log.Warn("indexDelete ZRemRangeByScores %s:%v,err:%v", index.NameKey, ToString(idInt), err)
				return err
			}
		case IndexType_UnSupport:
			log.Error("indexDelete unsupport index type")
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

//当前数据是否已经存在，存在则返回主键ID，唯一索引的才需要判断是否存在！
func (e *Engine) indexIsExistData(table *Table, beanValue, reflectVal reflect.Value, cols ...string) (int64, error) {
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
				e.Printfln("!index.IsUnique break")
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
			val, err := e.redisClient.ZRangeByScore(index.NameKey, zRangeBy).Result()
			if err != nil {
				log.Warn("indexIsExistData ZRangeByScore%s:%v,err:%v", index.NameKey, score, err)
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
			pkOldId, err := e.redisClient.ZScore(index.NameKey, strings.Join(members, "&")).Result()
			if err != nil && err.Error() != "redis: nil" {
				log.Warn("indexIsExistData %s:%v,err:%v", index.NameKey, strings.Join(members, "&"), err)
				return 0, err
			} else {
				if int64(pkOldId) > 0 {
					return int64(pkOldId), nil
				}
			}
		case IndexType_UnSupport:
			log.Error("indexIsExistData unsupport index type")
		}
	}
	return 0, nil
}

//todo: no thread safety! watch?
func (e *Engine) indexUpdate(table *Table, beanValue, reflectVal reflect.Value, cols ...string) error {
	e.Printfln("indexUpdate:%s,%v", table.Name, table.IndexesMap)
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
				e.Printfln("indexUpdate ColsIsExistIndex:%v,cols:%V", index.IndexColumn, cols)
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
			e.Printfln("score:%s", ToString(score))
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
			_, err := e.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				log.Warn("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
				return err
			}
		case IndexType_IdScore:
			//remove old index
			_, err := e.redisClient.ZRemRangeByScores(index.NameKey, ToString(pkFieldValue.Int()), ToString(pkFieldValue.Int())).Result()
			if err != nil {
				log.Warn("IndexUpdate ZRemRangeByScores %s:%v,err:%v", index.NameKey, ToString(pkFieldValue.Int()), err)
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
			_, err = e.redisClient.ZAdd(index.NameKey, redisZ).Result()
			if err != nil {
				log.Warn("IndexUpdate %s:%v,err:%v", index.NameKey, redisZ, err)
				return err
			}
		case IndexType_UnSupport:
			log.Error("IndexUpdate unsupport index type")
		}
	}
	return nil
}

func (e *Engine) IndexReBuild(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, reflectVal)
	if err != nil {
		return err
	}

	e.indexDrop(table, table.PrimaryKey)

	var offset int64 = 0
	var limit int64 = 100

	searchCon := NewSearchCondition(IndexType_IdMember, ScoreMin, ScoreMax, table.PrimaryKey)
	for {
		idAry, err := e.indexRange(table, searchCon, offset, limit)
		if err != nil {
			return err
		}
		if len(idAry) == 0 {
			break
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
			err = e.indexUpdate(table, newBeanValue, reflect.Indirect(newBeanValue))
			if err != nil {
				e.Printfln("indexReBuild indexUpdate(%v) err:%v", newBeanValue, err)
			}
		}
		if len(idAry) < int(limit) {
			break
		}
	}
	return nil
}

func (e *Engine) indexDrop(table *Table, except ...string) error {
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
		_, err := e.redisClient.Del(keys...).Result()
		return err
	}
	return nil
}
