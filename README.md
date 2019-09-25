# redis_orm
Object Relational Mapping use redis as a relational database。

#### 产出背景
```text
项目的快速迭代，不仅需要敏捷的开发，还需具备较高性能的和稳定性，单纯用关系型数据库有瓶颈，然后在关系型数据库基础上加分布式缓存或者进程内缓存有增加了开发和维护成本，
刚好项目中在用Redis，就考虑基于Redis的Hash和SortedSet两个数据结构来设计类似关系型数据库的ORM。经过多个版本的迭代，现在已经实现了ORM的基本功能，在应用中发现维护和查看数据
不太方便，又开发了[工作台](https://github.com/weikaishio/redis_orm_workbench).
```
#### 功能列表
* 基于对象的增、删、改、查、统计
* 基于Map的增、删、改、查、统计(方便用在redis_orm_workbench)
* 支持动态创建表、删除表、创建索引、重建索引
* 支持可配置的自动同步到MySql数据库(一般为了更方便的查询统计所用)


#### 使用说明
* 模型定义的标签说明

```go
    TagIdentifier = "redis_orm"
    	//定义是否索引，索引名自动生成 e.g.fmt.Sprintf("%s%s_%s", KeyIndexPrefix, strings.ToLower(table.Name), strings.ToLower(columnName)),
    TagIndex = "index"
    	//唯一索引 hash和走zscore一样都是O(1) 针对IndexType_IdMember有效，IndexType_IdScore的索引本来就是唯一的~
    TagUniqueIndex = "unique"
    	/*
    			要支持一种查询条件就得增加一个索引，定义用&连接联合索引中的字段，e.g.Status&Uid
    			组合索引 字符串则唯一！集合数据结构决定;
    		    除非用int64,44或224或2222来存放Score，即44：前4个字节uint32和后4个字节uint32
    			1、id作为score, 可以组合但是member唯一，唯一查询可用
    			此情况下的组合索引，直接按顺序拼接即可
    
    			2、id作为member，同一个member只能有一个score，该字段类型必须是长整型，数值索引可用，可以查询范围
    			此情况下的组合索引，仅仅支持两整型字段，左边32位 右边32位，支持范围查询的放左边
    	*/
    TagCombinedindex = "combinedindex"
	//默认值
	TagDefaultValue = "dft"
	//是否主键
	TagPrimaryKey = "pk"
	//自增~ 应用于主键
	TagAutoIncrement = "autoincr"
	//配置在主键的tag上，配置了该tag才能生效，同步到数据库
	TagSync2DB = "sync2db"
	//备注名
	TagComment = "comment"
	//是否支持自动写创建时间
	TagCreatedAt = "created_at"
	//是否支持自动写更新时间
	TagUpdatedAt = "updated_at"
```

* 模型例子

```go
type Faq struct {
	Id         int64  `redis_orm:"pk autoincr sync2db comment 'ID'"` //主键 自增 备注是ID
	Unique     int64  `redis_orm:"unique comment '唯一'"` //唯一索引
	Type       uint32 `redis_orm:"dft 1 comment '类型'"` //默认值：1
	Title      string `redis_orm:"dft 'faqtitle' index comment '标题'"` //默认值 faqtitle, 索引，备注 标题
	Content    string `redis_orm:"dft 'cnt' comment '内容'"`
	Hearts     uint32 `redis_orm:"dft 10 comment '点赞数'"`
	CreatedAt  int64  `redis_orm:"created_at comment '创建时间'"` //入库时自动写创建时间
	UpdatedAt  int64  `redis_orm:"updated_at comment '修改时间'"`
	TypeTitle  string `redis_orm:"combinedindex Type&Title comment '组合索引(类型&标题)'"` //组合索引 用到Type和Title两字段，字符串类型的索引，所以是唯一索引 
	TypeHearts int64  `redis_orm:"combinedindex Type&Hearts comment '组合索引(类型&赞数)'"` //组合索引 非唯一索引
}
```
* 需要引用的库、初始化方式等

```go
import (   
	"github.com/mkideal/log"
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
	//注：已省略redisClient可用性检测
	
	engine := redis_orm.NewEngine(redisClient)
	engine.IsShowLog(true)
	
	driver := "mysql"
	host := "127.0.0.1:3306"
	database := "bg_db"
	username := "root"
	password := ""
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&&allowOldPasswords=1&parseTime=true", username, password, host, database)
	
	dbEngine, err := xorm.NewEngine(driver, dataSourceName)
	if err != nil {
		log.Error("NewEngine:%s,err:%v", dataSourceName, err)
		return
    }
	//给redisORM对象设置同步到dbEngine数据库对象，每90秒同步一次
    engine.SetSync2DB(dbEngine, 90)
	//退出之前会执行同步到DB
	defer engine.Quit()
	
	faq := &models.Faq{
		Type: 1,
		Title:  "Title",
		Unique: 111,
		Content: "Content",
	}
	//插入数据
	engine.Insert(faq)
	
	//查询指定Id的数据
	model := &models.Faq{
		Id: 1,
   	}
	has, err := engine.Get(model)
   	if err != nil {
   		log.Error("Get(%d) err:%v", model.Id, err)
   		return
    }
    	
	//查询指定条件的数据
	searchCon := NewSearchConditionV2(faq.Unique, faq.Unique, 111)
	var ary []models.Faq
	count, err := engine.Find(0, 100, searchCon, &ary)
	if err != nil {
   		log.Error("Find(%v) err:%v", searchCon, err)
		return 
    }
	//其他请见engine_curd.go、engine_curd_by_map.go里面的方法....更新、删除等功能, 也可以看目录下面的测试代码
}
```
* 查看数据 

```text
建议使用配套的redis_orm_workbench来管理，可以维护表结构、数据和索引，方便直接在上面新增、修改和删除行数据。
也可以直接用redis-cli来查看数据，前缀tb:和ix:分别查看表数据和索引。

```