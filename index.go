package redis_orm

type IndexType int

const (
	IndexType_UnSupport IndexType = 0
	IndexType_IdMember  IndexType = 1
	IndexType_IdScore   IndexType = 2
)

type Index struct {
	NameKey  string
	ColumnName []string
	Type  IndexType
}

type SearchCondition struct {
	SearchColumn   string
	IndexType     IndexType
	FieldMaxValue interface{}
	FieldMinValue interface{}
}


