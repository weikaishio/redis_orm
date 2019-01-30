# redis_orm
[![Build Status](https://travis-ci.org/weikaishio/redis_orm.svg?branch=master)](https://travis-ci.org/weikaishio/redis_orm)
Object Relational Mapping use redis as a relational database

### started
####  Model Definition
* the tags's meaning
```go
    TagIdentifier = "redis_orm"
	//定义是否索引，索引名自动生成 e.g.fmt.Sprintf("%s%s_%s", KeyIndexPrefix, strings.ToLower(table.Name), strings.ToLower(columnName)),
	TagIndex = "index"
	//唯一索引 针对IndexType_IdMember有效，IndexType_IdScore的索引本来就是唯一的~
	TagUniqueIndex = "unique"
	/*
		要支持一种查询条件就得增加一个索引，用&连接联合索引中的字段
		组合索引 唯一！暂无法解决，除非用int64,前4个字节和后4个字节
		1、id作为score, 可以组合但是member唯一，唯一查询可用
		此情况下的组合索引，直接按顺序拼接即可

		2、该字段类型必须是长整型，id作为member，同一个member只能有一个score，数值索引可用，可以查询范围
		此情况下的组合索引，仅仅支持两整型字段，左边32位 右边32位，支持范围查询的放左边
	*/
	TagCombinedindex = "combinedindex"
	//默认值
	TagDefaultValue = "dft"
	//是否主键
	TagPrimaryKey = "pk"
	//自增~ 暂只支持主键
	TagAutoIncrement = "autoincr"
	//备注名
	TagComment = "comment"
	//是否支持自动写创建时间
	TagCreatedAt = "created_at"
	//是否支持自动写更新时间
	TagUpdatedAt = "updated_at"
```

* model sample
```cgo
type Faq struct {
	Id         int64  `redis_orm:"pk autoincr comment 'ID'"` //主键 自增 备注是ID
	Unique     int64  `redis_orm:"unique comment '唯一'"` //唯一索引
	Type       int    `redis_orm:"dft 1 comment '类型'"` //默认值：1
	Title      string `redis_orm:"dft 'faqtitle' index comment '标题'"` //默认值 faqtitle, 索引，备注 标题
	Content    string `redis_orm:"dft 'cnt' comment '内容'"`
	Hearts     int    `redis_orm:"dft 10 comment '点赞数'"`
	CreatedAt  int64  `redis_orm:"created_at comment '创建时间'"` //入库时自动写创建时间
	UpdatedAt  int64  `redis_orm:"updated_at comment '修改时间'"`
	TypeTitle  string `redis_orm:"combinedindex Type&Title comment '组合索引(类型&标题)'"` //组合索引 用到Type和Title两字段，字符串类型的索引，所以是唯一索引 
	TypeHearts int64  `redis_orm:"combinedindex Type&Hearts comment '组合索引(类型&赞数)'"` //组合索引 非唯一索引
}
```

#### samples
```go

import (  
	"github.com/go-redis/redis" 

	"github.com/weikaishio/redis_orm"
	"github.com/weikaishio/redis_orm/test/models"
)


func main() {
	options := redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	}

	redisClient := redis.NewClient(&options)

	engine := redis_orm.NewEngine(redisClient)
	engine.IsShowLog(true)
	faq := &models.Faq{
		Title:  "index3",
		Unique: time.Now().Unix(),
	}
	engine.Insert(faq)
}
    	
```
