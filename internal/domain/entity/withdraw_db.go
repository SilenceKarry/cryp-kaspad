package entity

import "github.com/shopspring/decimal"

type Withdraw struct {
	ID int64 `gorm:"column:id"`

	CreateTime int64 `gorm:"column:create_time"`
	UpdateTime int64 `gorm:"column:update_time"`

	MerchantType int `gorm:"column:merchant_type"`

	TxHash     string `gorm:"column:tx_hash"`
	CryptoType string `gorm:"column:crypto_type"`
	ChainType  string `gorm:"column:chain_type"`

	FromAddress string `gorm:"column:from_address"`
	ToAddress   string `gorm:"column:to_address"`

	Nonce uint64 `gorm:"column:nonce"`

	Amount     decimal.Decimal `gorm:"column:amount"`
	Memo       string          `gorm:"column:memo"`
	HasRetried int64           `gorm:"column:has_retried"`
	HasChain   int64           `gorm:"column:has_chain"`
}

func (w *Withdraw) TableName() string {
	return "withdraw"
}
