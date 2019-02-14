package table_from_ast

import (
	"testing"
	"encoding/json"
)

func TestTableFromAst(t *testing.T) {
	tableAry, err := TableFromAst("../example/models/faq.go")
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	for _, table := range tableAry {
		valTb,_:=json.Marshal(table)
		t.Logf("%v\n", string(valTb))
		//for _, col := range table.ColumnsMap {
		//	val,_:=json.Marshal(col)
		//	t.Logf("col:%v", string(val))
		//}
	}
}
