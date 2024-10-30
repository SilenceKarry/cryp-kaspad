package response

import (
	"cryp-kaspad/internal/utils"
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

// ShouldBindJSON 取得 Request body 資料，JSON 資料轉行為 struct
func ShouldBindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		var messages string
		if _, ok := err.(validator.ValidationErrors); ok {
			for _, err2 := range err.(validator.ValidationErrors) {
				messages += err2.Translate(translator) + ", "
			}

			// 過濾最後字尾 ", "
			messages = messages[:len(messages)-2]

		} else {
			messages = err.Error()
		}

		return errors.New(messages)
	}

	return nil
}

type Basic struct {
	Status Status `json:"status"`

	Data interface{} `json:"data"`
}

func NewSuccess(data interface{}) Basic {
	return Basic{
		Status: codeOk,
		Data:   data,
	}
}

func NewError(status Status) Basic {
	return Basic{
		Status: status,
		Data:   []string{},
	}
}

// Pagination 查詢分頁欄位
type Pagination struct {
	PageIndex int `json:"page_index" binding:"required"`
	PageSize  int `json:"page_size" binding:"required"`

	Total int64 `json:"total"`
}

func (p *Pagination) GetOffset() int {
	if p.PageIndex == utils.DefaultNotAssignInt || p.PageIndex == 0 {
		return 0
	}

	return (p.PageIndex - 1) * p.PageSize
}

func (p *Pagination) GetRowCount() int {
	if p.PageSize == utils.DefaultNotAssignInt || p.PageSize == 0 {
		return 25
	}

	return p.PageSize
}
