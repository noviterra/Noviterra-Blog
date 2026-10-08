package gormrepo

import (
	"Noviterra/server/app/model"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type RepoConfig struct {
	UserName string
	Password string
	Host     string
	Port     int
	DBName   string
}

func NewDB(repoconfig RepoConfig) *gorm.DB {
	dbStr := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		repoconfig.UserName,
		repoconfig.Password,
		repoconfig.Host,
		repoconfig.Port,
		repoconfig.DBName,
	)

	db, err := gorm.Open(mysql.Open(dbStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("数据库连接失败:%v", err))
	}

	// 数据库迁移
	err = db.AutoMigrate(&model.UserAccount{}, &model.UserInfo{}, &model.UserRefreshToken{})
	if err != nil {
		panic(fmt.Errorf("数据库表迁移失败:%v", err))
	}

	return db
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(repoConfig RepoConfig) *Repository {
	return &Repository{
		db: NewDB(repoConfig),
	}
}
