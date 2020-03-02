package services

import (
	"errors"
	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type IUserService interface {
	IsPwdSuccess(userName string, pwd string) (user *datamodels.User, isOk bool)
	InsertUser(user *datamodels.User) error
}

type UserService struct {
	UserRepository repositories.IUser
}

func NewUserService(repository repositories.IUser) IUserService {
	return &UserService{repository}
}


func (u *UserService) IsPwdSuccess(userName string, pwd string) (user *datamodels.User, isOk bool) {
	user, err := u.UserRepository.SelectByName(userName)
	if err != nil {
		return
	}

	isOk, _ = ValidatePassword(pwd, user.Password)
	if !isOk {
		return &datamodels.User{}, false
	}
	
	return
}

func (u *UserService) InsertUser(user *datamodels.User) error {
	pwdByte, errPwd := GeneratePassword(user.Password)
	if errPwd != nil {
		return errPwd
	}

	user.Password = string(pwdByte)
	return u.UserRepository.Insert(user)
}

func GeneratePassword(userPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

func ValidatePassword(userPassword string, hashed string) (isOK bool, err error) {
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassword)); err != nil {
		return false, errors.New("密码比对错误！")
	}
	return true, nil
}


//设置用户登录cookie
func SetLoginCookie(ctx iris.Context, user *datamodels.User, duration time.Duration) {
	uid := strconv.Itoa(int(user.Id))
	expireAt := strconv.Itoa(int(duration.Seconds()) + int(time.Now().Unix()))
	sign := MakeCookieSignMd5(uid, expireAt)

	sign = uid + "." + expireAt + "." + sign
	sign = url.QueryEscape(sign)

	ctx.SetCookieKV("sign", sign, iris.CookieExpires(duration))
}

//检查用户登录cookie
func CheckLoginCookie(ctx iris.Context) bool {
	sign := ctx.GetCookie("sign")
	signArr := strings.Split(sign, ".")

	if len(signArr) == 3 && signArr[2] == MakeCookieSignMd5(signArr[0], signArr[1]) {
		expireAt, err := strconv.Atoi(signArr[1])
		if err != nil {
			return false
		}

		if int64(expireAt) <= time.Now().Unix() {
			return false
		}

		return true
	} else {
		return false
	}
}


func MakeCookieSignMd5(uid, seconds string) string {
	sign := uid + "." + seconds + "abc123"
	return common.StringMd5(sign)
}

func GetUidFromCookie(ctx iris.Context) (uid int64, err error) {
	sign := ctx.GetCookie("sign")
	signArr := strings.Split(sign, ".")
	uidInt, err := strconv.Atoi(signArr[0])

	return int64(uidInt), err
}

