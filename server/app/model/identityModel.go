package model

import "time"

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
