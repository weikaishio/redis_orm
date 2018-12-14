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
	IsCascade       bool
	EnumOptions     map[string]int
	SetOptions      map[string]int
	Comment         string
	Type            reflect.Type
}

func NewEmptyColumn(colName string) *Column {
	return &Column{Name: colName, IsPrimaryKey: false,
		IsAutoIncrement: false}
}
