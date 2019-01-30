package table_from_ast

import "testing"

func TestTableFromAst(t *testing.T) {
	tableAry, err := TableFromAst("../example/models/faq.go")
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	for _, table := range tableAry {
		t.Logf("%v\n", *table)
	}
}
