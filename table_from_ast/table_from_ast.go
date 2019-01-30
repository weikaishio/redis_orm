package table_from_ast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"github.com/weikaishio/redis_orm"
)

func TableFromAst(fileName string) ([]*redis_orm.Table, error) {
	fileSet := token.NewFileSet()
	astF, err := parser.ParseFile(fileSet, fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var tableAry []*redis_orm.Table
	for _, decl := range astF.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid decl: %T", decl))
		}
		for _, ispec := range genDecl.Specs {
			switch spec := ispec.(type) {
			case *ast.TypeSpec:
				structType, ok := spec.Type.(*ast.StructType)
				if !ok {
					fmt.Printf("unsupported expr of TypeSpec: %T,ispec:%v,type:%v\n", spec.Type, ispec, spec.Type)
					continue
				}
				table := redis_orm.NewEmptyTable()
				table.Name = spec.Name.Obj.Name

				for seq, field := range structType.Fields.List {
					var (
						name = ""
					)
					if len(field.Names) == 0 {
						fmt.Printf("field.Type:%v\n", field.Type)
						continue
					} else if len(field.Names) == 1 {
						name = field.Names[0].Name
					} else {
						fmt.Printf("unsupported field.Names.length=%d\n", len(field.Names))
						continue
					}
					if field.Tag != nil {
						tag := field.Tag.Value
						if len(tag) > 0 && tag[0] == '`' {
							tag = tag[1:]
						}
						if len(tag) > 0 && tag[len(tag)-1] == '`' {
							tag = tag[:len(tag)-1]
						}
						rdsTagStr := reflect.StructTag(tag).Get(redis_orm.TagIdentifier)
						if len(rdsTagStr) > 0 {
							tmpStrs := strings.SplitN(rdsTagStr, ",", 2)
							if len(tmpStrs) > 0 {
								rdsTagStr = tmpStrs[0]
							}
						}
						//todo:field.Type -> reflect.Kind
						redis_orm.MapTag(table, seq, name, reflect.String, rdsTagStr)
					}
				}
				tableAry = append(tableAry, table)
			default:
				fmt.Printf("unsupported spec: %T\n", ispec)
			}
		}
	}
	return tableAry, nil
}
