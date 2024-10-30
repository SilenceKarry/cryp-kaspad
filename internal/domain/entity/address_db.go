package entity

type Address struct {
	ID int64 `gorm:"column:id"`

	CreateTime int64 `gorm:"column:create_time"`
	UpdateTime int64 `gorm:"column:update_time"`

	MerchantType int    `gorm:"column:merchant_type"`
	Address      string `gorm:"column:address"`

	ChainType string `gorm:"column:chain_type"`
}

func (addr *Address) TableName() string {
	return "address"
}
