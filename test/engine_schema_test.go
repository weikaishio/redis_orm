package test

import (
	"fmt"
	"github.com/weikaishio/redis_orm"
	"github.com/weikaishio/redis_orm/example/models"
	"reflect"
	"strings"
	"testing"
)

func TestEngine_CreateTable(t *testing.T) {
	faq := &models.Faq{}
	err := engine.Schema.CreateTable(faq)
	t.Logf("CreateTable(%v),err:%v", faq, err)
}

func TestEngine_GetTable(t *testing.T) {
	faq := &models.Faq{
		Title:   "为啥",
		Content: "我也不知道",
		Hearts:  20,
	}
	val := reflect.ValueOf(faq)
	table, err := engine.GetTableByReflect(val, reflect.Indirect(val))
	t.Logf("table:%v,err:%v", table, err)
}

func TestEngine_SchemaColumnsFromColumn(t *testing.T){
	table,has:=engine.Tables["faq"]
	if !has{
		t.Logf("!has table")
		return
	}

	for _,column:=range table.ColumnsMap{
		t.Logf("table.column:%v",column)
		if column.EnumOptions!=nil{
			t.Logf("EnumOptions:%v",column.EnumOptions)
			var enumAry []string
			for k, _ := range column.EnumOptions {
				enumAry = append(enumAry, k)
			}
			t.Logf("tags:%v",fmt.Sprintf("%s '%s'", redis_orm.TagEnum, strings.Join(enumAry, ",")))
		}
	}
}