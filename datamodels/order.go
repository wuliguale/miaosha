package datamodels

type Order struct {
	Id uint64 `json:"id" sql:"id"`
	Uid uint32 `json:"uid" sql:"uid"`
	Pid uint64 `json:"pid" sql:"pid"`
	State uint8 `json:"state" sql:"state"`
	CreateAt int64 `json:"createAt" sql:"create_at"`
	UpdateAt int64 `json:"updateAt" sql:"update_at"`
}

const (
	OrderWait    = iota
	OrderSuccess //1
	OrderFailed  //2
)

func (Order) TableName() string {
	return "miaosha_order"
}

