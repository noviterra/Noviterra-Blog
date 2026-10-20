package model

import "time"

type UserAccount struct {
	ID        string `gorm:"primaryKey;size:36"`
	Password  string `gorm:"size:128"`
	Email     string `gorm:"size:128"`
	Status    int    `gorm:"default:0;comment:账号状态 0正常 1封禁"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserAccount) TableName() string {
	return "account"
}

type UserRefreshToken struct {
	ID           string    `gorm:"primaryKey;size:36"`
	RefreshToken string    `gorm:"size:256;not null"`
	ExpireAt     time.Time `gorm:"not null"`
	CreateAt     time.Time
	UpdatedAt    time.Time
}

func (UserRefreshToken) TableName() string {
	return "auth"
}

type UserInfo struct {
	ID               string     `gorm:"primaryKey;size:36;comment:用户ID"`
	Name             string     `gorm:"size:64;not null;default:'';index;comment:昵称"`
	Email            string     `gorm:"size:128"`
	Age              int        `gorm:"default:0;comment:年龄"`
	Gender           int        `gorm:"default:0;comment:性别 0未知 1男 2女"`
	Avatar           string     `gorm:"size:255;default:'';comment:头像URL"`
	SelfIntroduction string     `gorm:"size:512;default:'';comment:个人简介"`
	Role             int        `gorm:"default:0;comment:用户身份 0普通 1管理员 2超级管理员（站主）"`
	IsDeleted        bool       `gorm:"default:false;index;comment:是否删除"`
	DeletedAt        *time.Time `gorm:"index;comment:删除时间"`
	CreatedAt        time.Time  `gorm:"comment:创建时间"`
	UpdatedAt        time.Time  `gorm:"comment:更新时间"`
}

func (UserInfo) TableName() string {
	return "users"
}
