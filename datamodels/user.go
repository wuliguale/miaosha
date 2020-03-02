package datamodels

type User struct {
	Id uint64  `json:"id" sql:"id"`
	NickName string `json:"nickName" sql:"nick_name"`
	UserName string `json:"userName" sql:"user_ame"`
	Password string `gorm:"column:pass_word" json:"-" sql:"pass_word"`
	State uint8 `json:"state" sql:"state"`
	CreateAt int64 `json:"createAt" sql:"create_at"`
	UpdateAt int64 `json:"updateAt" sql:"update_at"`
}

func (User) TableName() string {
	return "miaosha_user"
}
