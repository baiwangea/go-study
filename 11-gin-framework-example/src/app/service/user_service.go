package service

import (
	"errors"
	"gin-framework-example/src/app/model"
	"gin-framework-example/src/app/response"
	"gin-framework-example/src/pkg/db"
	"time"
)

func toUserResponse(user model.User) response.UserResponse {
	return response.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Mobile:    user.Mobile,
		Level:     user.Level,
		Status:    user.Status,
		CreatedAt: time.Unix(*user.CreateTime, 0),
	}
}

func CreateUser(user *model.User) (*response.UserResponse, error) {
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	userResponse := toUserResponse(*user)
	return &userResponse, nil
}

func GetUser(id uint) (*response.UserResponse, error) {
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	userResponse := toUserResponse(user)
	return &userResponse, nil
}

func UpdateUser(id uint, updatedUser *model.User) (*response.UserResponse, error) {
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}

	// Note: This only updates the Username. For a real-world application,
	// you would likely have a more robust update mechanism.
	user.Username = updatedUser.Username

	if err := db.DB.Save(&user).Error; err != nil {
		return nil, err
	}
	userResponse := toUserResponse(user)
	return &userResponse, nil
}

func DeleteUser(id uint) error {
	if err := db.DB.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

func Login(username, password string) (*response.UserResponse, string, error) {
	var user model.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", err
	}

	if user.Password != password {
		return nil, "", errors.New("invalid password")
	}

	userResponse := toUserResponse(user)
	return &userResponse, user.Token, nil
}
