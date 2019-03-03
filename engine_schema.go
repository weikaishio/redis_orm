package redis_orm

import (
	"fmt"
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
	*Engine
}

func NewSchemaEngine(e *Engine) *SchemaEngine {
	schemaEngine := &SchemaEngine{
		Engine: e,
	}
	var beans []interface{}
	beans = append(beans, &SchemaTablesTb{}, &SchemaColumnsTb{}, &SchemaIndexsTb{})
	for _, bean := range beans {
		beanValue := reflect.ValueOf(bean)
		beanIndirectValue := reflect.Indirect(beanValue)
		schemaEngine.GetTableByReflect(beanValue, beanIndirectValue)
	}
	return schemaEngine
}

func (s *SchemaEngine) CreateTableByTable(table *Table) error {
	tablesTb := SchemaTablesFromTable(table)
	err := s.Insert(tablesTb)
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

	affectedRows, err := s.InsertMulti(columnAry...)
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
	affectedRows, err = s.InsertMulti(indexAry...)
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return ERR_UnKnowError
	}
	s.tablesMutex.Lock()
	s.Tables[table.Name] = table
	s.tablesMutex.Unlock()

	return nil
}
func (s *SchemaEngine) CreateTable(bean interface{}) error {
	beanValue := reflect.ValueOf(bean)
	beanIndirectValue := reflect.Indirect(beanValue)

	table, has := s.GetTableByName(s.TableName(beanIndirectValue))
	if has {
		s.Printfln("CreateTable GetTableByName(%s),has", s.TableName(beanIndirectValue))
		return Err_DataHadAvailable
	}
	table, err := s.mapTable(beanIndirectValue)
	if err != nil {
		return err
	}
	//if table != nil {
	//	s.tablesMutex.Lock()
	//	s.Tables[table.Name] = table
	//	s.tablesMutex.Unlock()
	//}

	err = s.CreateTable(table)
	if err == nil && s.isSync2DB && table.IsSync2DB {
		s.syncDB.Create2DB(bean)
	}
	return err
}

/*
createTable by AST
ALTER TABLE table_name ADD COLUMN column_name string DEFAULT abc COMMENT 测试列 AFTER updated_at;

ALTER TABLE table_name ADD INDEX index_name uid;

ALTER TABLE table_name ADD INDEX index_name (uid, name);

CREATE TABLE table_name (
  column_name1 INT(11) COMMENT 列1,
  column_name2 string DEFAULT abc COMMENT 列2,
  column_name3 BIGINT(20) created_at COMMENT 添加时间,
) pk(column_name1) autoincr=1 sync2db COMMENT='xx表';
*/
func (s *SchemaEngine) AlterTable(sql string) error {
	return nil
}

//the bean is new, the column which it is the new need to be added
func (s *SchemaEngine) AddColumn(bean interface{}, cols ...string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	currentTable, err := s.mapTable(reflectVal)
	if err != nil {
		return err
	}
	table, has := s.Tables[currentTable.Name]
	if !has {
		return ERR_UnKnowTable
	}

	var isNeedRefreshTable bool

	for _, col := range cols {
		v, has := table.ColumnsMap[col]
		if has {
			s.Printfln("AddColumn(%s) col:%s, already exist", table.Name, col)
			continue
		}
		v, has = currentTable.ColumnsMap[col]
		if !has {
			s.Printfln("AddColumn(%s) col:%s not exist at new table bean", table.Name, col)
			continue
		}
		columnsTb := SchemaColumnsFromColumn(table.TableId, v)
		err := s.Insert(columnsTb)
		if err != nil {
			return err
		} else {
			isNeedRefreshTable = true
		}
		break
	}

	if isNeedRefreshTable {
		s.ReloadTable(table.Name)
		if s.isSync2DB && table.IsSync2DB {
			s.syncDB.Create2DB(bean)
		}
	}
	return nil
}
func (s *SchemaEngine) RemoveColumn(bean interface{}, cols ...string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	currentTable, err := s.mapTable(reflectVal)
	if err != nil {
		return err
	}
	table, has := s.Tables[currentTable.Name]
	if !has {
		return ERR_UnKnowTable
	}

	var isNeedRefreshTable bool

	for _, col := range cols {
		_, has = table.ColumnsMap[col]
		if !has {
			s.Printfln("RemoveColumn(%s) col:%s, not exist", table.Name, col)
			continue
		}
		_, err = s.DeleteByCondition(&SchemaColumnsTb{}, NewSearchConditionV2(fmt.Sprintf("%d&%s", table.TableId, col), fmt.Sprintf("%d&%s", table.TableId, col), "TableId", "ColumnName"))
		if err != nil {
			return err
		} else {
			isNeedRefreshTable = true
		}
	}

	if isNeedRefreshTable {
		s.ReloadTable(table.Name)
		if s.isSync2DB && table.IsSync2DB {
			s.syncDB.Create2DB(bean)
		}
	}
	return nil
}
func (s *SchemaEngine) AddIndex(bean interface{}, cols ...string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	currentTable, err := s.mapTable(reflectVal)
	if err != nil {
		return err
	}
	table, has := s.Tables[currentTable.Name]
	if !has {
		return ERR_UnKnowTable
	}

	var isNeedRefreshTable bool

	for _, col := range cols {
		v, has := table.IndexesMap[table.GetIndexKey(col)]
		if has {
			s.Printfln("AddIndex(%s) index:%s, already exist", table.Name, col)
			continue
		}

		v, has = currentTable.IndexesMap[table.GetIndexKey(col)]
		if !has {
			s.Printfln("AddIndex(%s) index:%s not exist at new table bean", table.Name, col)
			continue
		}

		index := SchemaIndexsFromColumn(table.TableId, v)
		err := s.Insert(index)
		if err != nil {
			return err
		} else {
			s.Engine.Index.ReBuild(bean)
			isNeedRefreshTable = true
		}
		break
	}

	if isNeedRefreshTable {
		s.ReloadTable(table.Name)
		if s.isSync2DB && table.IsSync2DB {
			s.syncDB.Create2DB(bean)
		}
	}
	return nil
}
func (s *SchemaEngine) RemoveIndex(bean interface{}, cols ...string) error {
	beanValue := reflect.ValueOf(bean)
	reflectVal := reflect.Indirect(beanValue)
	currentTable, err := s.mapTable(reflectVal)
	if err != nil {
		return err
	}
	table, has := s.Tables[currentTable.Name]
	if !has {
		return ERR_UnKnowTable
	}

	var isNeedRefreshTable bool

	for _, col := range cols {
		dropIndex, has := table.IndexesMap[table.GetIndexKey(col)]
		if !has {
			s.Printfln("AddIndex(%s) index:%s, not exist", table.Name, col)
			continue
		}
		_, err = s.DeleteByCondition(&SchemaIndexsTb{}, NewSearchConditionV2(fmt.Sprintf("%d&%s", table.TableId, col), fmt.Sprintf("%d&%s", table.TableId, col), "TableId", "IndexName"))
		if err != nil {
			return err
		} else {
			s.Engine.Index.DropSingleIndex(dropIndex)
			isNeedRefreshTable = true
		}
	}

	if isNeedRefreshTable {
		s.ReloadTable(table.Name)
		if s.isSync2DB && table.IsSync2DB {
			s.syncDB.Create2DB(bean)
		}
	}
	return nil
}
func (s *SchemaEngine) TableDrop(table *Table) error {
	affectedRow, err := s.DeleteByCondition(&SchemaTablesTb{}, NewSearchConditionV2(table.Name, table.Name, "TableName"))
	if err != nil {
		s.Printfln("Delete Table err:%v", err)
	}
	if affectedRow == 0 {
		//return Err_DataNotAvailable
		s.Printfln("Delete Table: table not available")
	}

	affectedRow, err = s.DeleteByCondition(&SchemaColumnsTb{}, NewSearchConditionV2(table.TableId, table.TableId, "TableId"))
	if err != nil {
		s.Printfln("Delete Column err:%v", err)
	}
	if affectedRow == 0 {
		//return Err_DataNotAvailable
		s.Printfln("Delete Column: column not available")
	}

	_, err = s.DeleteByCondition(&SchemaIndexsTb{}, NewSearchConditionV2(table.TableId, table.TableId, "TableId"))
	if err != nil {
		return err
	}

	//TableTruncate 敏感操作，需要独立调用，等可以支持动态改表结构，无需使用TableDrop再恢复
	//err = s.TableTruncate(bean)
	//if err == nil {
	s.tablesMutex.Lock()
	delete(s.Tables, table.Name)
	s.tablesMutex.Unlock()
	//}
	return err
}

func (s *SchemaEngine) ShowTables() []string {
	s.tablesMutex.RLock()
	defer s.tablesMutex.RUnlock()
	tableAry := make([]string, 0)
	for _, v := range s.Tables {
		if !strings.Contains(NeedMapTable, v.Name) {
			tableAry = append(tableAry, v.Name)
		}
	}
	return tableAry
}

func (s *SchemaEngine) ReloadTable(tableName string) (*Table, error) {
	schemaTable := &SchemaTablesTb{}
	has, err := s.GetByCondition(schemaTable, NewSearchConditionV2(tableName, tableName, "TableName"))
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ERR_UnKnowTable
	}

	table := TableFromSchemaTables(schemaTable)

	var columnsAry []*SchemaColumnsTb
	_, err = s.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &columnsAry)
	if err != nil {
		s.Printfln("SchemaTables2MapTables(%v) find SchemaColumnsTb,err:%v", schemaTable, err)
		return nil, err
	}
	for _, schemaColumn := range columnsAry {
		table.ColumnsSeq = append(table.ColumnsSeq, schemaColumn.ColumnName)
		table.ColumnsMap[schemaColumn.ColumnName] = ColumnFromSchemaColumns(schemaColumn, schemaTable)
	}

	var indexsAry []*SchemaIndexsTb
	_, err = s.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &indexsAry)
	if err != nil {
		s.Printfln("SchemaTables2MapTables(%v) find SchemaIndexsTb,err:%v", schemaTable, err)
		return nil, err
	}
	for _, schemaIndex := range indexsAry {
		table.IndexesMap[strings.ToLower(schemaIndex.IndexColumn)] = IndexFromSchemaIndexs(schemaIndex)
	}

	s.tablesMutex.Lock()
	s.Tables[table.Name] = table
	s.tablesMutex.Unlock()
	return table, nil
}
func (s *SchemaEngine) ReloadTables() ([]*Table, error) {
	tables := make([]*Table, 0)
	var tablesAry []*SchemaTablesTb
	count, err := s.Find(0, int64(math.MaxInt64), NewSearchConditionV2(ScoreMin, ScoreMax, "Id"), &tablesAry)
	if err != nil {
		return tables, err
	}
	if count != int64(len(tablesAry)) {
		s.Printfln("ReloadTables count:%d !=len(tablesAry):%d", count, len(tablesAry))
		return tables, ERR_UnKnowError
	}

	for _, schemaTable := range tablesAry {
		table := TableFromSchemaTables(schemaTable)

		var columnsAry []*SchemaColumnsTb
		_, err := s.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &columnsAry)
		if err != nil {
			s.Printfln("SchemaTables2MapTables(%v) find SchemaColumnsTb,err:%v", schemaTable, err)
			continue
		}
		for _, schemaColumn := range columnsAry {
			table.ColumnsSeq = append(table.ColumnsSeq, schemaColumn.ColumnName)
			table.ColumnsMap[schemaColumn.ColumnName] = ColumnFromSchemaColumns(schemaColumn, schemaTable)
		}

		var indexsAry []*SchemaIndexsTb
		_, err = s.Find(0, int64(math.MaxInt64), NewSearchConditionV2(schemaTable.Id, schemaTable.Id, "TableId"), &indexsAry)
		if err != nil {
			s.Printfln("SchemaTables2MapTables(%v) find SchemaIndexsTb,err:%v", schemaTable, err)
			continue
		}
		for _, schemaIndex := range indexsAry {
			table.IndexesMap[strings.ToLower(schemaIndex.IndexColumn)] = IndexFromSchemaIndexs(schemaIndex)
		}
		tables = append(tables, table)
	}
	if len(tables) > 0 {
		for _, table := range tables {
			s.Tables[table.Name] = table
		}
	}
	return tables, nil
}
