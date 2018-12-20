package redis_orm

import (
	"reflect"
)

type Column struct {
	Name            string
	DefaultValue    string
	IsPrimaryKey    bool
	IsAutoIncrement bool
	IsCreated       bool
	IsUpdated       bool
	IsCombinedIndex bool //it 's only used for judge wherther need insert or delete and so on
	IsCascade       bool
	EnumOptions     map[string]int
	SetOptions      map[string]int
	Comment         string
	Type            reflect.Type//only support base type
}

func NewEmptyColumn(colName string) *Column {
	return &Column{Name: colName, IsPrimaryKey: false,
		IsAutoIncrement: false}
}
