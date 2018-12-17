package redis_orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

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
func (table *Table) AddIndex(typ reflect.Type, columnName string) {
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

	index := &Index{
		NameKey:    fmt.Sprintf("%s%s_%s", KeyIndexPrefix, strings.ToLower(table.Name), strings.ToLower(columnName)),
		ColumnName: []string{columnName},
		Type:       indexType,
	}
	table.mutex.Lock()
	table.IndexesMap[strings.ToLower(columnName)] = index
	table.mutex.Unlock()
}
func (table *Table) AddColumn(col *Column) {
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
