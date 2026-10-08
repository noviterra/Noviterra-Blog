package gormrepo

import (
	"Noviterra/server/app/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) CheckIDExists(id string) (bool, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Select("id").Where("id = ?", id).Take((&u)).Error
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CheckEmailExists(email string) (bool, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Select("email").Where("email = ?", email).Take((&u)).Error
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) InsertUsersAccount(userAccount model.UserAccount, userInfo model.UserInfo) error {
	// 使用事务保证两张表结果保持一致
	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *Repository) UpdateUsersAccount(userAccount model.UserAccount) error {
	err := r.db.Model(&model.UserAccount{}).Where("id = ?", userAccount.ID).Updates(userAccount).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateUsersInfo(userInfo model.UserInfo) error {
	err := r.db.Model(&model.UserInfo{}).Where("id = ?", userInfo.ID).Updates(userInfo).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetUserPasswordByID(id string) (string, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.Password, nil
}

func (r *Repository) GetUserIDByEmail(email string) (string, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Where("email = ?", email).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.ID, nil
}

func (r *Repository) GetUserEmailByID(id string) (string, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.Email, nil
}

func (r *Repository) GetUserInfoByID(id string) (*model.UserInfo, error) {
	var userInfo model.UserInfo
	err := r.db.Model(&model.UserInfo{}).Where("id = ?", id).Take(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userInfo, nil
}

// 存储RefreshToken Create->Upset 有则更新 无则创建
func (r *Repository) InsertUserToken(userToken model.UserRefreshToken) error {
	err := r.db.
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

// 更新Token信息
func (r *Repository) UpdateUserToken(userToken model.UserRefreshToken) error {
	err := r.db.Model(&model.UserRefreshToken{}).Where("id = ?", userToken.ID).Updates(userToken).Error
	if err != nil {
		return err
	}
	return nil
}

// 获取用户账号状态
func (r *Repository) GetUserAccountStatus(id string) (int, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return -1, nil
		}
		return -1, err
	}
	return u.Status, nil
}

// 获取用户身份
func (r *Repository) GetUserRole(id string) (int, error) {
	var u model.UserInfo
	err := r.db.Model(&model.UserInfo{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return -1, nil
		}
		return -1, err
	}
	return u.Role, nil
}

// 获取用户全部信息
func (r *Repository) GetAllUsersInfo(size, page int) ([]model.UserInfo, error) {
	var userList []model.UserInfo

	// 分页查询
	err := r.db.Offset((page - 1) * size).Limit(size).Find(&userList).Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}

// 获取用户密码
func (r *Repository) GetUserPassword(id string) (string, error) {
	var u model.UserAccount
	err := r.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.Password, nil
}
