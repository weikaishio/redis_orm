package redis_orm

import (
	"math"
	"reflect"
	"strings"
)

/*

todo:DB隔离, DB如何兼容已有的Table，暂不用吧，redis有自己的DB

Done:存表、字段、索引结构

todo:逆向生成模型

todo:改表结构？需要存一个版本号~ pub/sub, 修改了表结构需要reload table, schemaTable -> mapTable
增加，修改，删除字段，有索引的会自动删除索引
增加，修改，删除索引，重建索引

*/
func (e *Engine) CreateTable(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)
	e.tablesMutex.RLock()
	table, ok := e.Tables[e.tbName(beanIndirectValue)]
	e.tablesMutex.RUnlock()
	if ok {
		return Err_DataHadAvailable
	}
	table, err := e.mapTable(beanIndirectValue)
	if err != nil {
		return err
	}
	if table != nil {
		e.tablesMutex.Lock()
		e.Tables[table.Name] = table
		e.tablesMutex.Unlock()
	}
	tablesTb := SchemaTablesFromTable(table)
	err = e.Insert(tablesTb)
	if err != nil {
		if err != nil {
			return err
		}
	}

	columnAry := make([]interface{}, 0)
	for _, v := range table.ColumnsMap {
		columnsTb := SchemaColumnsFromColumn(tablesTb.Id, v)
		columnAry = append(columnAry, columnsTb)
	}
	affectedRows, err := e.InsertMulti(columnAry...)
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return ERR_UnKnowError
	}
	indexAry := make([]interface{}, 0)
	for _, v := range table.IndexesMap {
		indexsTb := SchemaIndexsFromColumn(tablesTb.Id, v)
		indexAry = append(indexAry, indexsTb)
	}
	affectedRows, err = e.InsertMulti(indexAry...)
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return ERR_UnKnowError
	}
	e.tablesMutex.Lock()
	e.Tables[table.Name] = table
	e.tablesMutex.Unlock()
	return nil
}

func (e *Engine) TableDrop(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, beanIndirectValue)
	if err != nil {
		return err
	}
	//tablesTb := SchemaTablesFromTable(table)
	affectedRow, err := e.DeleteByCondition(&SchemaTablesTb{}, NewSearchConditionV2(table.Name, table.Name, "TableName"))
	if err != nil {
		return err
	}
	if affectedRow == 0 {
		return Err_DataNotAvailable
	}

	affectedRow, err = e.DeleteByCondition(&SchemaColumnsTb{}, NewSearchConditionV2(table.TableId, table.TableId, "TableId"))
	if err != nil {
		return err
	}
	if affectedRow == 0 {
		return Err_DataNotAvailable
	}

	_, err = e.DeleteByCondition(&SchemaIndexsTb{}, NewSearchConditionV2(table.TableId, table.TableId, "TableId"))
	if err != nil {
		return err
	}

	err = e.TableTruncate(bean)
	if err == nil {
		e.tablesMutex.Lock()
		delete(e.Tables, table.Name)
		e.tablesMutex.Unlock()
	}
	return err
}

func (e *Engine) ShowTables() []string {
	e.tablesMutex.RLock()
	defer e.tablesMutex.RUnlock()
	tableAry := make([]string, 0)
	for _, v := range e.Tables {
		if !strings.Contains(NeedMapTable, v.Name) {
			tableAry = append(tableAry, v.Name)
		}
	}
	return tableAry
}

func (e *Engine) ReloadTables() ([]*Table, error) {
	tables := make([]*Table, 0)
	var tablesAry []*SchemaTablesTb
	count, err := e.Find(0, int64(math.MaxInt64), NewSearchConditionV2(ScoreMin, ScoreMax, "Id"), &tablesAry)
	if err != nil {
		return tables, err
	}
	if count != int64(len(tablesAry)) {
		e.Printfln("ReloadTables count:%d !=len(tablesAry):%d", count, len(tablesAry))
		return tables, ERR_UnKnowError
	}

	for _, schemaTable := range tablesAry {
		table := TableFromSchemaTables(schemaTable)

		var columnsAry []*SchemaColumnsTb
		_, err := e.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &columnsAry)
		if err != nil {
			e.Printfln("SchemaTables2MapTables(%v) find SchemaColumnsTb,err:%v", schemaTable, err)
			continue
		}
		for _, schemaColumn := range columnsAry {
			table.ColumnsSeq = append(table.ColumnsSeq, schemaColumn.ColumnName)
			table.ColumnsMap[schemaColumn.ColumnName] = ColumnFromSchemaColumns(schemaColumn, schemaTable)
		}

		var indexsAry []*SchemaIndexsTb
		_, err = e.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &indexsAry)
		if err != nil {
			e.Printfln("SchemaTables2MapTables(%v) find SchemaIndexsTb,err:%v", schemaTable, err)
			continue
		}
		for _, schemaIndex := range indexsAry {
			table.IndexesMap[strings.ToLower(schemaIndex.IndexColumn)] = IndexFromSchemaIndexs(schemaIndex)
		}
		tables = append(tables, table)
	}
	if len(tables) > 0 {
		for _, table := range tables {
			e.Tables[table.Name] = table
		}
	}
	return tables, nil
}
