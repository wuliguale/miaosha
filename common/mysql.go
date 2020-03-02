package common

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//创建mysql 连接
func NewMysqlConn() (db *gorm.DB, err error) {
	db, err = gorm.Open("mysql", "root:root@(192.168.125.128:3306)/miaosha?charset=utf8&parseTime=True&loc=Local")
	//defer db.Close()
	return db, err
}


