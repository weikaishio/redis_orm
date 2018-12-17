package redis_orm

import "strings"

type IndexType int

const (
	IndexType_UnSupport IndexType = 0
	IndexType_IdMember  IndexType = 1
	IndexType_IdScore   IndexType = 2
)

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
