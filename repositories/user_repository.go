package repositories

import (
	"errors"
	"github.com/jinzhu/gorm"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
)

type IUser interface {
	Conn() error
	Insert(user *datamodels.User) error
	SelectByName(userName string) (user *datamodels.User, err error)
	SelectByPk(uid uint64) (user *datamodels.User, err error)
}

type UserRepository struct {
	mysqlConn *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUser {
	return &UserRepository{mysqlConn: db}
}


func (u *UserRepository) Conn() (err error) {
	if u.mysqlConn == nil {
		mysql, errMysql := common.NewMysqlConn()
		if errMysql != nil {
			return errMysql
		}
		u.mysqlConn = mysql
	}
	return nil
}

func (u *UserRepository) SelectByPk(uid uint64) (user *datamodels.User, err error) {
	user = &datamodels.User{}
	res := u.mysqlConn.First(user, uid)
	return user, res.Error
}


func (u *UserRepository) SelectByName(userName string) (user *datamodels.User, err error) {
	user = &datamodels.User{}

	if userName == "" {
		return user, errors.New("userName empty")
	}

	err = u.Conn()
	if err != nil {
		return user, err
	}

	res := u.mysqlConn.Where("user_name = ?", userName).First(user)
	if gorm.IsRecordNotFoundError(res.Error) {
		return user, errors.New("user not exist")
	}

	return user, res.Error
}

func (u *UserRepository) Insert(user *datamodels.User) error {
	return u.mysqlConn.Create(user).Error
}
