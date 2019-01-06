package redis_orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

/*
SET @table_schema='employees';
SELECT
    table_name,
    table_type,
    engine,
    table_rows,
    avg_row_length,
    data_length,
    index_length,
    table_collation,
    create_time
FROM
    information_schema.tables
WHERE
    table_schema = @table_schema
ORDER BY table_name;
*/
type SchemaTablesTb struct {
	Id            int64  `redis_orm:"pk autoincr comment 'ID'"`
	TableName     string `redis_orm:"unique comment '唯一'"`
	TableComment  string `redis_orm:"dft '' index comment '表注释'"` //暂时没用上
	PrimaryKey    string `redis_orm:"comment '主键字段'"`
	AutoIncrement string `redis_orm:"comment '自增字段'"`
	Created       string `redis_orm:"comment '创建时间字段'"`
	Updated       string `redis_orm:"comment '更新时间字段'"`
	CreatedAt     int64  `redis_orm:"created_at comment '创建时间'"`
	UpdatedAt     int64  `redis_orm:"updated_at comment '修改时间'"`
}

func SchemaTablesFromTable(table *Table) *SchemaTablesTb {
	return &SchemaTablesTb{
		TableName:     table.Name,
		TableComment:  table.Name,
		PrimaryKey:    table.PrimaryKey,
		AutoIncrement: table.AutoIncrement,
		Created:       table.Created,
		Updated:       table.Updated,
	}
}

type Table struct {
	Name          string
	Type          reflect.Type
	ColumnsSeq    []string
	ColumnsMap    map[string]*Column
	IndexesMap    map[string]*Index
	PrimaryKey    string
	AutoIncrement string
	Created       string
	Updated       string
	mutex         sync.RWMutex
}

func TableFromSchemaTables(table *SchemaTablesTb) *Table {
	return &Table{
		Name:          table.TableName,
		PrimaryKey:    table.PrimaryKey,
		AutoIncrement: table.AutoIncrement,
		ColumnsMap:    make(map[string]*Column),
		IndexesMap:    make(map[string]*Index),
		Created:       table.Created,
		Updated:       table.Updated,
	}
}

func NewEmptyTable() *Table {
	return &Table{Name: "", Type: nil,
		ColumnsSeq: make([]string, 0),
		ColumnsMap: make(map[string]*Column),
		IndexesMap: make(map[string]*Index),
	}
}
func (table *Table) GetAutoIncrKey() string {
	if table.AutoIncrement != "" {
		return fmt.Sprintf("%s%s", KeyAutoIncrPrefix, strings.ToLower(table.AutoIncrement))
	} else {
		return ""
	}
}
func (table *Table) GetTableKey() string {
	return fmt.Sprintf("%s%s", KeyTbPrefix, strings.ToLower(table.Name))
}
func (table *Table) AddIndex(typ reflect.Type, indexColumn, columnName, comment string, isUnique bool) {
	var indexType IndexType
	switch typ.Kind() {
	case reflect.String:
		indexType = IndexType_IdScore

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8:
		fallthrough
	case reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		indexType = IndexType_IdMember

	case reflect.Uintptr, reflect.Ptr:
		fallthrough
	case reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Interface, reflect.Map:
		fallthrough
	case reflect.Slice, reflect.Struct, reflect.Bool, reflect.UnsafePointer:
		fallthrough
	default:
		indexType = IndexType_UnSupport
	}

	if indexType == IndexType_UnSupport {
		return
	}
	index := &Index{
		NameKey:     table.GetIndexKey(columnName),
		IndexColumn: strings.Split(indexColumn, "&"),
		Comment:     comment,
		Type:        indexType,
		IsUnique:    isUnique,
	}
	table.mutex.Lock()
	table.IndexesMap[strings.ToLower(indexColumn)] = index
	table.mutex.Unlock()
}
func (table *Table) GetIndexKey(col string) string {
	return fmt.Sprintf("%s%s_%s", KeyIndexPrefix, strings.ToLower(table.Name), strings.ToLower(col))
}
func (table *Table) AddColumn(col *Column) {
	if col.IsCombinedIndex {
		return
	}
	table.ColumnsSeq = append(table.ColumnsSeq, col.Name)
	colName := col.Name
	table.ColumnsMap[colName] = col

	if col.IsPrimaryKey {
		table.PrimaryKey = col.Name
	}
	if col.IsAutoIncrement {
		table.AutoIncrement = col.Name
	}
	if col.IsCreated {
		table.Created = col.Name
	}
	if col.IsUpdated {
		table.Updated = col.Name
	}
}
