package entity

type Config struct {
	ID int64 `gorm:"column:id"`

	Key        string `gorm:"column:key"`
	Value      string `gorm:"column:value"`
	CreateTime int64  `gorm:"column:create_time"`
	UpdateTime int64  `gorm:"column:update_time"`
}

func (bh *Config) TableName() string {
	return "config"
}
