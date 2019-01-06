package redis_orm

import (
	"math"
	"reflect"
	"strings"
)

/*

todo:DB隔离, DB如何兼容已有的Table，暂不用吧，redis有自己的DB

todo:存表、字段、索引结构 ok 待测试

todo:改表结构？需要存一个版本号~ pub/sub, 修改了表结构需要reload table, schemaTable -> mapTable

todo:逆向生成模型
*/
func (e *Engine) CreateTable(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, beanIndirectValue)
	if err != nil {
		return err
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
	return nil
}

func (e *Engine) TableDrop(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, beanIndirectValue)
	if err != nil {
		return err
	}
	tablesTb := SchemaTablesFromTable(table)
	affectedRow, err := e.DeleteByCondition(tablesTb, NewSearchConditionV2(table.Name, table.Name, "TableName"))
	if err != nil {
		return err
	}
	if affectedRow == 0 {
		return Err_DataNotAvailable
	}

	err = e.TableTruncate(bean)
	return err
}

//keys tb:*
func (e *Engine) ShowTables() []string {
	e.tablesMutex.RLock()
	defer e.tablesMutex.RUnlock()
	tableAry := make([]string, 0)
	for _, v := range e.Tables {
		tableAry = append(tableAry, v.Name)
	}
	return tableAry
}

func (e *Engine) ReloadTables() ([]*SchemaTablesTb, error) {
	bean := &SchemaTablesTb{}
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)
	table, err := e.GetTable(beanValue, beanIndirectValue)
	if err != nil {
		return nil, err
	}
	var tablesAry []*SchemaTablesTb
	count, err := e.Find(0, int64(math.MaxInt64), NewSearchConditionV2(ScoreMin, ScoreMax, table.PrimaryKey), &tablesAry)
	if err != nil {
		return nil, err
	}
	if count != int64(len(tablesAry)) {
		e.Printfln("ReloadTables count:%d !=len(tablesAry):%d", count, len(tablesAry))
		return nil, ERR_UnKnowError
	}
	return tablesAry, nil
}
func (e *Engine) SchemaTables2MapTables() ([]*Table, error) {
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
	return tables, nil
}
