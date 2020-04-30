package repositories

import (
	"errors"
	"github.com/jinzhu/gorm"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
)

type IUser interface {
	Insert(user *datamodels.User) error
	SelectByName(userName string) (user *datamodels.User, err error)
	SelectByPk(uid uint64) (user *datamodels.User, err error)
}

type UserRepository struct {
	mysqlPool *common.MysqlPool
}

func NewUserRepository(mysqlPool *common.MysqlPool) IUser {
	return &UserRepository{mysqlPool:mysqlPool}
}

func (u *UserRepository) SelectByPk(uid uint64) (user *datamodels.User, err error) {
	db, err := u.mysqlPool.Get()
	defer u.mysqlPool.Put(db)

	if err != nil {
		return nil, err
	}

	user = &datamodels.User{}
	res := db.First(user, uid)
	return user, res.Error
}


func (u *UserRepository) SelectByName(userName string) (user *datamodels.User, err error) {
	user = &datamodels.User{}

	if userName == "" {
		return user, errors.New("userName empty")
	}

	db, err := u.mysqlPool.Get()
	defer u.mysqlPool.Put(db)
	if err != nil {
		return nil, err
	}

	res := db.Where("user_name = ?", userName).First(user)
	if gorm.IsRecordNotFoundError(res.Error) {
		return user, errors.New("user not exist")
	}

	return user, res.Error
}

func (u *UserRepository) Insert(user *datamodels.User) error {
	db, err := u.mysqlPool.Get()
	defer u.mysqlPool.Put(db)
	if err != nil {
		return err
	}

	return db.Create(user).Error
}
