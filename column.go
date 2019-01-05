package redis_orm

import (
	"reflect"
)

type SchemaColumnsTb struct {
	Id            int64  `redis_orm:"pk autoincr comment 'ID'"`
	TableId       int64  `redis_orm:"index comment '表ID'"`
	ColumnName    string `redis_orm:"comment '字段名'"`
	ColumnComment string `redis_orm:"dft '' index comment '字段注释'"`
	DataType      string `redis_orm:"comment '数据类型'"`
	DefaultValue  string `redis_orm:"comment '默认值'"`
	CreatedAt     int64  `redis_orm:"created_at comment '创建时间'"`
	UpdatedAt     int64  `redis_orm:"updated_at comment '修改时间'"`
}

func SchemaColumnsFromColumn(tableId int64, v *Column) *SchemaColumnsTb {
	return &SchemaColumnsTb{
		TableId:       tableId,
		ColumnName:    v.Name,
		ColumnComment: v.Comment,
		DefaultValue:  v.DefaultValue,
		DataType:      v.Type.Kind().String(),
	}
}

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
	Type            reflect.Type //only support base type
}

func NewEmptyColumn(colName string) *Column {
	return &Column{Name: colName, IsPrimaryKey: false,
		IsAutoIncrement: false}
}
