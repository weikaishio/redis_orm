package redis_orm

import "strings"

type IndexType int

const (
	IndexType_UnSupport IndexType = 0
	IndexType_IdMember  IndexType = 1
	IndexType_IdScore   IndexType = 2
)

type IndexsTb struct {
	Id           int64  `redis_orm:"pk autoincr comment 'ID'"`
	TableId      int    `redis_orm:"index comment '表ID'"`
	IndexName    string `redis_orm:"comment '索引名'"`
	IndexComment string `redis_orm:"dft '' index comment '索引注释'"`
	IndexColumn  string `redis_orm:"comment '索引字段，&分割'"`
	IndexType    int    `redis_orm:"comment '数据类型'"`
	IsUnique     bool   `redis_orm:"comment '是否唯一索引'"`
	CreatedAt    int64  `redis_orm:"created_at comment '创建时间'"`
	UpdatedAt    int64  `redis_orm:"updated_at comment '修改时间'"`
}

type Index struct {
	NameKey     string
	IndexColumn []string
	Type        IndexType
	IsUnique    bool
}

type SearchCondition struct {
	SearchColumn  []string
	IndexType     IndexType
	FieldMaxValue interface{}
	FieldMinValue interface{}
	IsAsc         bool
}

func NewSearchCondition(indexType IndexType, minVal, maxVal interface{}, column ...string) *SearchCondition {
	return &SearchCondition{
		SearchColumn:  column,
		IndexType:     indexType,
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
