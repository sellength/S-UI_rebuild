package service

import (
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/logger"
	"s-ui/util/common"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
}

const bcryptPasswordPrefix = "$2"

func (s *UserService) GetFirstUser() (*model.User, error) {
	db := database.GetDB()

	user := &model.User{}
	err := db.Model(model.User{}).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateFirstUser(username string, password string) error {
	if username == "" {
		return common.NewError("username can not be empty")
	} else if password == "" {
		return common.NewError("password can not be empty")
	}
	hashedPassword, err := hashLoginPassword(password)
	if err != nil {
		return err
	}
	db := database.GetDB()
	user := &model.User{}
	err = db.Model(model.User{}).First(user).Error
	if database.IsNotFound(err) {
		user.Username = username
		user.Password = hashedPassword
		return db.Model(model.User{}).Create(user).Error
	} else if err != nil {
		return err
	}
	user.Username = username
	user.Password = hashedPassword
	return db.Save(user).Error
}

func (s *UserService) Login(username string, password string, remoteIP string) (string, error) {
	user := s.CheckUser(username, password, remoteIP)
	if user == nil {
		return "", common.NewError("wrong user or password! IP: ", remoteIP)
	}
	return user.Username, nil
}

func (s *UserService) CheckUser(username string, password string, remoteIP string) *model.User {
	db := database.GetDB()

	user := &model.User{}
	err := db.Model(model.User{}).
		Where("username = ?", username).
		First(user).
		Error
	if err == gorm.ErrRecordNotFound {
		return nil
	} else if err != nil {
		logger.Warning("check user err:", err, " IP: ", remoteIP)
		return nil
	}
	if !verifyLoginPassword(user.Password, password) {
		return nil
	}
	if shouldUpgradeLoginPassword(user.Password) {
		hashedPassword, err := hashLoginPassword(password)
		if err != nil {
			logger.Warning("unable to hash login password", err)
			return nil
		}
		if err := db.Model(user).Update("password", hashedPassword).Error; err != nil {
			logger.Warning("unable to upgrade login password hash", err)
			return nil
		}
		user.Password = hashedPassword
	}

	lastLoginTxt := time.Now().Format("2006-01-02 15:04:05") + " " + remoteIP
	err = db.Model(model.User{}).
		Where("username = ?", username).
		Update("last_logins", &lastLoginTxt).Error
	if err != nil {
		logger.Warning("unable to log login data", err)
	}
	return user
}

func (s *UserService) GetUsers() (*[]model.User, error) {
	var users []model.User
	db := database.GetDB()
	err := db.Model(model.User{}).Select("id,username,last_logins").Scan(&users).Error
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (s *UserService) ChangePass(id string, oldPass string, newUser string, newPass string) error {
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).Where("id = ?", id).First(user).Error
	if err != nil || database.IsNotFound(err) {
		return err
	}
	if !verifyLoginPassword(user.Password, oldPass) {
		return gorm.ErrRecordNotFound
	}
	hashedPassword, err := hashLoginPassword(newPass)
	if err != nil {
		return err
	}
	user.Username = newUser
	user.Password = hashedPassword
	return db.Save(user).Error
}

func hashLoginPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", common.NewError("password can not be empty")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func verifyLoginPassword(storedPassword string, inputPassword string) bool {
	if strings.HasPrefix(storedPassword, bcryptPasswordPrefix) {
		return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(inputPassword)) == nil
	}
	return storedPassword == inputPassword
}

func shouldUpgradeLoginPassword(storedPassword string) bool {
	return storedPassword != "" && !strings.HasPrefix(storedPassword, bcryptPasswordPrefix)
}
