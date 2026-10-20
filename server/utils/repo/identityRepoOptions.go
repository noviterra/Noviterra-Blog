package repo

import (
	"Noviterra/server/app/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IdentityRepository struct {
	db *gorm.DB
}

func NewIdentityRespository(repoConfig RepoConfig, field string) *IdentityRepository {
	return &IdentityRepository{
		db: NewDB(repoConfig, field),
	}
}

func (ir *IdentityRepository) InsertUsersAccount(userAccount model.UserAccount, userInfo model.UserInfo) error {
	// 使用事务保证两张表结果保持一致
	return ir.db.Transaction(func(tx *gorm.DB) error {
		// 插入用户信息表
		err := tx.Create(&userInfo).Error
		if err != nil {
			return err
		}

		// 插入账号表
		err = tx.Create(&userAccount).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (ir *IdentityRepository) GetUserIDByEmail(email string) (string, error) {
	var u model.UserAccount
	err := ir.db.Model(&model.UserAccount{}).Where("email = ?", email).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.ID, nil
}

func (ir *IdentityRepository) GetUserEmailByID(id string) (string, error) {
	var u model.UserAccount
	err := ir.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.Email, nil
}

func (ir *IdentityRepository) UpdateUsersAccount(userAccount model.UserAccount) error {
	err := ir.db.Model(&model.UserAccount{}).Where("id = ?", userAccount.ID).Updates(userAccount).Error
	if err != nil {
		return err
	}
	return nil
}

// 存储RefreshToken Create->Upset 有则更新 无则创建
func (ir *IdentityRepository) InsertUserToken(userToken model.UserRefreshToken) error {
	err := ir.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}}, // 主键
			DoUpdates: clause.AssignmentColumns([]string{
				"refresh_token", "expire_at", "updated_at", // 更新字段 不更新created_at
			}),
		}).
		Create(userToken).Error

	if err != nil {
		return err
	}
	return nil
}

func (ir *IdentityRepository) GetUserPasswordByID(id string) (string, error) {
	var u model.UserAccount
	err := ir.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.Password, nil
}

func (ir *IdentityRepository) CheckIDExists(id string) (bool, error) {
	var u model.UserAccount
	err := ir.db.Model(&model.UserAccount{}).Select("id").Where("id = ?", id).Take((&u)).Error
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ir *IdentityRepository) CheckEmailExists(email string) (bool, error) {
	var u model.UserAccount
	err := ir.db.Model(&model.UserAccount{}).Select("email").Where("email = ?", email).Take((&u)).Error
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// 更新Token信息
func (ir *IdentityRepository) UpdateUserToken(userToken model.UserRefreshToken) error {
	err := ir.db.Model(&model.UserRefreshToken{}).Where("id = ?", userToken.ID).Updates(userToken).Error
	if err != nil {
		return err
	}
	return nil
}

// 获取用户账号状态
func (ir *IdentityRepository) GetUserAccountStatus(id string) (int, error) {
	var u model.UserAccount
	err := ir.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return -1, nil
		}
		return -1, err
	}
	return u.Status, nil
}

// 获取用户身份
func (ir *IdentityRepository) GetUserRole(id string) (int, error) {
	var u model.UserInfo
	err := ir.db.Model(&model.UserInfo{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return -1, nil
		}
		return -1, err
	}
	return u.Role, nil
}
