package datamodels

type Product struct {
	Id uint64  `json:"id" sql:"id"`
	ProductName string `json:"productName" sql:"product_name"`
	Num uint32  `json:"productNum" sql:"num"`
	Image string `json:"productImage" sql:"image"`
	Url string `json:"productUrl" sql:"url"`
	State uint8 `json:"state" sql:"state"`
	CreateAt int64 `json:"createAt" sql:"create_at"`
	UpdateAt int64 `json:"updateAt" sql:"update_at"`
}


func (p *Product) TableName() string {
	return "miaosha_product"
}
