package service

import (
	"errors"
	"gin-framework-example/internal/app/model"
	"gin-framework-example/pkg/db"
)

func CreateUser(user *model.User) (*model.User, error) {
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetUser(id uint) (*model.User, error) {
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(id uint, updatedUser *model.User) (*model.User, error) {
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
	return &user, nil
}

func DeleteUser(id uint) error {
	if err := db.DB.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

func Login(username, password string) (*model.User, error) {
	var user model.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	if user.Password != password {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}
