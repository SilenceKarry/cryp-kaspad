package entity

import "github.com/shopspring/decimal"

type Tokens struct {
	ID int64 `gorm:"column:id"`

	CreateTime int64 `gorm:"column:create_time"`
	UpdateTime int64 `gorm:"column:update_time"`

	CryptoType   string `gorm:"column:crypto_type"`
	ChainType    string `gorm:"column:chain_type"`
	ContractAddr string `gorm:"column:contract_addr"`

	Decimals       int             `gorm:"column:decimals"`
	GasLimit       int64           `gorm:"column:gas_limit"`
	GasPrice       decimal.Decimal `gorm:"column:gas_price"`
	TransactionFee decimal.Decimal `gorm:"column:transaction_fee"`

	ContractAbi string `gorm:"column:contract_abi"`
}

func (t *Tokens) TableName() string {
	return "tokens"
}
