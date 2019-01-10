package models

type Faq struct {
	Id         int64  `redis_orm:"pk autoincr sync2db comment 'ID'"`
	Unique     int64  `redis_orm:"unique comment '唯一'"`
	Type       int    `redis_orm:"index dft 1 comment '类型'"`
	Title      string `redis_orm:"dft 'faqtitle' index comment '标题'"`
	Content    string `redis_orm:"dft 'cnt' comment '内容'"`
	Hearts     int    `redis_orm:"index dft 10 comment '点赞数'"`
	CreatedAt  int64  `redis_orm:"created_at comment '创建时间'"`
	UpdatedAt  int64  `redis_orm:"updated_at comment '修改时间'"`
	TypeTitle  string `redis_orm:"combinedindex Type&Title comment '组合索引(类型&标题)'"`
	TypeHearts int64  `redis_orm:"combinedindex Type&Hearts comment '组合索引(类型&赞数)'"`
}

/*
CREATE TABLE `bg_db`.`faq_tb` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `type` TINYINT(4) NULL,
  `title` VARCHAR(45) NULL DEFAULT 'faqtitle',
  `content` VARCHAR(200) NULL DEFAULT 'cnt',
  `hearts` INT(11) NULL DEFAULT 20,
  `created_at` BIGINT(20) NULL,
  `update_at` BIGINT(20) NULL,
  PRIMARY KEY (`id`),
  INDEX `ix_type_hearts` (`type` ASC, `hearts` ASC));
*/


type FaqTb struct {
	Id        int64  `redis_orm:"pk autoincr sync2db comment 'ID'" xorm:"pk autoincr BIGINT(20)"`
	Type      int    `redis_orm:"index dft 1 comment '类型'" xorm:"default 1 TINYINT(4)"`
	Title     string `redis_orm:"dft 'faqtitle' index comment '标题'" xorm:"default 'faqtitle' VARCHAR(64)"`
	Content   string `redis_orm:"dft 'cnt' comment '内容'" xorm:"default 'cnt' VARCHAR(200)"`
	Hearts    int    `redis_orm:"index dft 10 comment '点赞数'" xorm:"default 20 INT(11)"`
	CreatedAt int64  `redis_orm:"created_at comment '创建时间'" xorm:"-> created not null BIGINT(20)"`
	UpdatedAt int64  `redis_orm:"updated_at comment '修改时间'" xorm:"-> updated not null BIGINT(20)"`
}