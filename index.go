package redis_orm

import "strings"

type IndexType int

const (
	IndexType_UnSupport IndexType = 0
	IndexType_IdMember  IndexType = 1
	IndexType_IdScore   IndexType = 2 //todo:sortedset -> hash
)

type SchemaIndexsTb struct {
	Id               int64  `redis_orm:"pk autoincr comment 'ID'"`
	TableId          int64  `redis_orm:"index comment '表ID'"`
	Seq              byte   `redis_orm:"comment '索引顺序'"`
	IndexName        string `redis_orm:"comment '索引名'"`
	IndexComment     string `redis_orm:"dft '' comment '索引注释'"`
	IndexColumn      string `redis_orm:"comment '索引字段，&分割'"`
	IndexType        int    `redis_orm:"comment '数据类型'"`
	IsUnique         bool   `redis_orm:"comment '是否唯一索引'"`
	TableIdIndexName string `redis_orm:"combinedindex TableId&IndexName comment '组合索引(表ID&索引名)'"`
	CreatedAt        int64  `redis_orm:"created_at comment '创建时间'"`
	UpdatedAt        int64  `redis_orm:"updated_at comment '修改时间'"`
}

func SchemaIndexsFromColumn(tableId int64, v *Index) *SchemaIndexsTb {
	return &SchemaIndexsTb{
		TableId:      tableId,
		Seq:          v.Seq,
		IndexName:    v.NameKey,
		IndexComment: v.Comment,
		IndexColumn:  strings.Join(v.IndexColumn, "&"),
		IsUnique:     v.IsUnique,
		IndexType:    int(v.Type),
	}
}

type Index struct {
	Seq         byte
	NameKey     string
	IndexColumn []string
	Comment     string
	Type        IndexType
	IsUnique    bool
}

func IndexFromSchemaIndexs(v *SchemaIndexsTb) *Index {
	column := &Index{
		NameKey:     v.IndexName,
		Seq:         v.Seq,
		Comment:     v.IndexComment,
		IndexColumn: strings.Split(v.IndexColumn, "&"),
		IsUnique:    v.IsUnique,
		Type:        IndexType(v.IndexType),
	}
	return column
}

type SearchCondition struct {
	SearchColumn []string
	//IndexType     IndexType
	FieldMaxValue interface{}
	FieldMinValue interface{}
	IsAsc         bool
}

//NewSearchCondition will be deprecated, please use NewSearchConditionV2 instead
func NewSearchCondition(indexType IndexType, minVal, maxVal interface{}, column ...string) *SearchCondition {
	return &SearchCondition{
		SearchColumn:  column,
		FieldMaxValue: maxVal,
		FieldMinValue: minVal,
	}
}

func NewSearchConditionV2(minVal, maxVal interface{}, column ...string) *SearchCondition {
	return &SearchCondition{
		SearchColumn:  column,
		FieldMaxValue: maxVal,
		FieldMinValue: minVal,
	}
}
func (s *SearchCondition) IsEqualIndexName(index *Index) bool {
	return strings.ToLower(strings.Join(s.SearchColumn, "&")) == strings.ToLower(strings.Join(index.IndexColumn, "&"))
}
func (s *SearchCondition) Name() string {
	return strings.Join(s.SearchColumn, "&")
}
