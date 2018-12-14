package redis_orm

import (
	"errors"
)

const (
	TagIdentifier    = "rds"
	TagIndex         = "index"
	TagDefaultValue  = "dft"
	TagPrimaryKey    = "pk"
	TagAutoIncrement = "autoincr"
	TagComment       = "comment"
	TagCreatedAt     = "created_at"
	TagUpdatedAt     = "updated_at"

	KeyTbPrefix       = "tb:"
	KeyIndexPrefix    = "ix:"
	KeyAutoIncrPrefix = "autoincr_last_"
)

var (
	ERR_UnKnowField           = errors.New("redis-orm-error:unknow field")
	ERR_UnKnowError           = errors.New("redis-orm-error:unknow error")
	ERR_NotSupportIndexField  = errors.New("redis-orm-error:not support this filed's index")
	Err_UnSupportedType       = errors.New("redis-orm-error:unsupported type")
	Err_FieldValueInvalid     = errors.New("redis-orm-error:field value invalid")
	Err_PrimaryKeyNotFound    = errors.New("redis-orm-error:primarykey not found")
	Err_PrimaryKeyTypeInvalid = errors.New("redis-orm-error:primarykey type invalid")
	Err_DataNotAvailable      = errors.New("redis-orm-error:data not exist")
)
