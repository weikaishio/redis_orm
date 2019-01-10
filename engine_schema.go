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
type SchemaEngine struct {
	engine *Engine
}

func NewSchemaEngine(e *Engine) *SchemaEngine {
	schemaEngine := &SchemaEngine{
		engine: e,
	}
	var beans []interface{}
	beans = append(beans, &SchemaTablesTb{}, &SchemaColumnsTb{}, &SchemaIndexsTb{})
	for _, bean := range beans {
		beanValue := reflect.ValueOf(bean)
		beanIndirectValue := reflect.Indirect(beanValue)
		schemaEngine.engine.GetTableByReflect(beanValue, beanIndirectValue)
	}
	return schemaEngine
}

func (s *SchemaEngine) CreateTable(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)
	s.engine.tablesMutex.RLock()
	table, ok := s.engine.Tables[s.engine.TableName(beanIndirectValue)]
	s.engine.tablesMutex.RUnlock()
	if ok {
		return Err_DataHadAvailable
	}
	table, err := s.engine.mapTable(beanIndirectValue)
	if err != nil {
		return err
	}
	if table != nil {
		s.engine.tablesMutex.Lock()
		s.engine.Tables[table.Name] = table
		s.engine.tablesMutex.Unlock()
	}
	tablesTb := SchemaTablesFromTable(table)
	err = s.engine.Insert(tablesTb)
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
	affectedRows, err := s.engine.InsertMulti(columnAry...)
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
	affectedRows, err = s.engine.InsertMulti(indexAry...)
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return ERR_UnKnowError
	}
	s.engine.tablesMutex.Lock()
	s.engine.Tables[table.Name] = table
	s.engine.tablesMutex.Unlock()
	return nil
}

/*
todo: AddColumn
*/
//the bean is new, the column which it is the new need to be added
func (s *SchemaEngine) AddColumn(bean interface{}, colName string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	_, err := s.engine.mapTable(reflectVal)
	if err != nil {
		return err
	}
	//for k,v:=range table.ColumnsMap{
	//	if k==colName {
	//		columnAry := make([]interface{}, 0)
	//		for _, v := range table.ColumnsMap {
	//			columnsTb := SchemaColumnsFromColumn(tablesTb.Id, v)
	//			columnAry = append(columnAry, columnsTb)
	//		}
	//		affectedRows, err := s.engine.InsertMulti(columnAry...)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}
	return nil
}
func (s *SchemaEngine) RemoveColumn(bean interface{}, colName string) error {
	return nil
}
func (s *SchemaEngine) AddIndex(bean interface{}, colName string) error {
	return s.AddColumn(bean, colName)
}
func (s *SchemaEngine) RemoveIndex(bean interface{}, colName string) error {
	return s.RemoveColumn(bean, colName)
}
func (s *SchemaEngine) TableDrop(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)

	table, has := s.engine.GetTableByName(s.engine.TableName(beanIndirectValue))
	if !has {
		return ERR_UnKnowTable
	}

	//tablesTb := SchemaTablesFromTable(table)
	affectedRow, err := s.engine.DeleteByCondition(&SchemaTablesTb{}, NewSearchConditionV2(table.Name, table.Name, "TableName"))
	if err != nil {
		return err
	}
	if affectedRow == 0 {
		return Err_DataNotAvailable
	}

	affectedRow, err = s.engine.DeleteByCondition(&SchemaColumnsTb{}, NewSearchConditionV2(table.TableId, table.TableId, "TableId"))
	if err != nil {
		return err
	}
	if affectedRow == 0 {
		return Err_DataNotAvailable
	}

	_, err = s.engine.DeleteByCondition(&SchemaIndexsTb{}, NewSearchConditionV2(table.TableId, table.TableId, "TableId"))
	if err != nil {
		return err
	}

	err = s.engine.TableTruncate(bean)
	if err == nil {
		s.engine.tablesMutex.Lock()
		delete(s.engine.Tables, table.Name)
		s.engine.tablesMutex.Unlock()
	}
	return err
}

func (s *SchemaEngine) ShowTables() []string {
	s.engine.tablesMutex.RLock()
	defer s.engine.tablesMutex.RUnlock()
	tableAry := make([]string, 0)
	for _, v := range s.engine.Tables {
		if !strings.Contains(NeedMapTable, v.Name) {
			tableAry = append(tableAry, v.Name)
		}
	}
	return tableAry
}

func (s *SchemaEngine) ReloadTables() ([]*Table, error) {
	tables := make([]*Table, 0)
	var tablesAry []*SchemaTablesTb
	count, err := s.engine.Find(0, int64(math.MaxInt64), NewSearchConditionV2(ScoreMin, ScoreMax, "Id"), &tablesAry)
	if err != nil {
		return tables, err
	}
	if count != int64(len(tablesAry)) {
		s.engine.Printfln("ReloadTables count:%d !=len(tablesAry):%d", count, len(tablesAry))
		return tables, ERR_UnKnowError
	}

	for _, schemaTable := range tablesAry {
		table := TableFromSchemaTables(schemaTable)

		var columnsAry []*SchemaColumnsTb
		_, err := s.engine.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &columnsAry)
		if err != nil {
			s.engine.Printfln("SchemaTables2MapTables(%v) find SchemaColumnsTb,err:%v", schemaTable, err)
			continue
		}
		for _, schemaColumn := range columnsAry {
			table.ColumnsSeq = append(table.ColumnsSeq, schemaColumn.ColumnName)
			table.ColumnsMap[schemaColumn.ColumnName] = ColumnFromSchemaColumns(schemaColumn, schemaTable)
		}

		var indexsAry []*SchemaIndexsTb
		_, err = s.engine.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &indexsAry)
		if err != nil {
			s.engine.Printfln("SchemaTables2MapTables(%v) find SchemaIndexsTb,err:%v", schemaTable, err)
			continue
		}
		for _, schemaIndex := range indexsAry {
			table.IndexesMap[strings.ToLower(schemaIndex.IndexColumn)] = IndexFromSchemaIndexs(schemaIndex)
		}
		tables = append(tables, table)
	}
	if len(tables) > 0 {
		for _, table := range tables {
			s.engine.Tables[table.Name] = table
		}
	}
	return tables, nil
}
