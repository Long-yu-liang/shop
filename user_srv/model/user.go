package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32     `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
	IsDeleted bool `gorm:"column:is_deleted"`
}

/*
 1. 密文保存密码 2.密文不可反解
    1.采用非对称加密
    2.md5 信息摘要算法
    密码如果不可以反解，用户找回密码
*/
type User struct {
	BaseModel
	Mobile   string     `gorm:"index:idx_mobile;unique;type:varchar(11);not null"`
	Password string     `gorm:"type:varchar(100);not null"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"`
	Gender   string     `gorm:"column:gender;default:male;type:varchar(6) comment 'femal 女 male 男'"`
	Role     int        `gorm:"column:role;default:1;type:int comment '1 普通用户 2 管理员'"`
}

func (User) TableName() string {
	return "user"
}
