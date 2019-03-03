package table_from_ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"encoding/json"
	"github.com/weikaishio/redis_orm"
)

/*
https://studygolang.com/articles/6709
https://github.com/yuroyoro/goast-viewer
*/
func TableFromAst(fileName string, fileContent string) ([]*redis_orm.Table, error) {
	fileSet := token.NewFileSet()
	astF, err := parser.ParseFile(fileSet, fileName, fileContent, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile err:%v", err)
	}
	var tableAry []*redis_orm.Table
	for _, decl := range astF.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			return nil, fmt.Errorf("invalid decl: %T", decl)
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
						//Done:field.Type -> reflect.Kind
						fmt.Printf("field.Type:%T,%s\n", field.Type, redis_orm.ToString(field.Type))
						fieldTypeStr := redis_orm.ToString(field.Type)
						var identObj ast.Ident
						err = json.Unmarshal([]byte(fieldTypeStr), &identObj)
						if err != nil {
							fmt.Printf("identObj unmarshal err:%v\n", err)
							continue
						}

						redis_orm.MapTableColumnFromTag(table, seq, name, identObj.Name, rdsTagStr)
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
