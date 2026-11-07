package repo

import (
	"Noviterra/server/app/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(repoConfig RepoConfig, field string) *UserRepository {
	return &UserRepository{
		db: NewDB(repoConfig, field),
	}
}

func (ur *UserRepository) UpdateUsersInfo(userInfo model.UserInfo) error {
	err := ur.db.Model(&model.UserInfo{}).Where("id = ?", userInfo.ID).Updates(userInfo).Error
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) GetUserInfoByID(id string) (*model.UserInfo, error) {
	var userInfo model.UserInfo
	err := ur.db.Model(&model.UserInfo{}).Where("id = ?", id).Take(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userInfo, nil
}

// 获取用户全部信息
func (ur *UserRepository) GetAllUsersInfo(size, page int) ([]model.UserInfo, error) {
	var userList []model.UserInfo

	// 分页查询
	err := ur.db.Offset((page - 1) * size).Limit(size).Find(&userList).Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}

// 获取用户密码
func (ur *UserRepository) GetUserPassword(id string) (string, error) {
	var u model.UserAccount
	err := ur.db.Model(&model.UserAccount{}).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.Password, nil
}

// GetUserInfoInComment 得到展示在评论区中的用户信息
func (ur *UserRepository) GetUserInfoInComment(ctx context.Context, userIDs []string) (map[string]model.CommentSpeaker, error) {
	if len(userIDs) == 0 {
		return map[string]model.CommentSpeaker{}, nil
	}

	var speakerInfo []model.CommentSpeaker
	if err := ur.db.WithContext(ctx).
		Model(&model.UserInfo{}).
		Select([]string{"id", "name", "avatar"}).
		Where("id IN ? AND is_deleted = ?", userIDs, false).
		Find(&speakerInfo).Error; err != nil {
		return nil, err
	}

	// 转换为map[<userID>]<UserInfo>
	result := make(map[string]model.CommentSpeaker, len(speakerInfo))
	for _, speaker := range speakerInfo {
		result[speaker.ID] = speaker
	}

	return result, nil
}
