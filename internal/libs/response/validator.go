package response

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	locale2FieldMap map[string]map[string]string

	locale = "en"

	translator ut.Translator
)

func RegisterValidator() error {
	locale2FieldMap = make(map[string]map[string]string)

	locale2FieldMap[locale] = newField2Name()

	e := en.New()
	uni := ut.New(e, e)
	trans, found := uni.GetTranslator(locale)
	if !found {
		return fmt.Errorf("expected '%t' Got '%t'", true, found)
	}

	studentGenderFunc := studentGenderFunc{}
	studentStatusFunc := studentStatusFunc{}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 註冊翻譯器
		err := enTranslations.RegisterDefaultTranslations(v, trans)
		if err != nil {
			panic(err)
		}

		// 註冊一個函式，獲取struct tag裡自定義的label作為欄位名
		v.RegisterTagNameFunc(registerTagNameFunc)

		// 註冊 studentGenderTag
		if err := v.RegisterValidation(studentGenderTag,
			studentGenderFunc.Validator); err != nil {
			panic(err)
		}

		v.RegisterTranslation(studentGenderTag, trans,
			studentGenderFunc.Translations, studentGenderFunc.Translation)

		// 註冊 studentStatusTag
		if err := v.RegisterValidation(studentStatusTag,
			studentStatusFunc.Validator); err != nil {
			panic(err)
		}

		v.RegisterTranslation(studentStatusTag, trans,
			studentStatusFunc.Translations, studentStatusFunc.Translation)
	}

	translator = trans

	return nil
}

// registerTagNameFunc 註冊欄位對應轉譯的文字
func registerTagNameFunc(fld reflect.StructField) string {
	fieldName := strings.ToLower(fld.Name)
	return locale2FieldMap[locale][fieldName]
}

func newField2Name() map[string]string {
	field2Name := make(map[string]string)

	field2Name["gender"] = "Field gender"
	field2Name["status"] = "Field status"

	return field2Name
}

const (
	statusEnable  = 1
	statusDisable = 0

	studentGenderTag = "studentGender"
	studentStatusTag = "studentStatus"
)

type studentGenderFunc struct {
}

// Validator 提供驗證資料正確性 func
func (ssf *studentGenderFunc) Validator(fl validator.FieldLevel) bool {
	if val, ok := fl.Field().Interface().(int); ok {
		if val == statusDisable || val == statusEnable {
			return true
		}
		return false
	}

	return true
}

// Translations 提供錯誤訊息格式
func (ssf *studentGenderFunc) Translations(ut ut.Translator) error {
	return ut.Add(studentGenderTag, "{0} only woman(1) or man(0)", true)
}

// Translation 提供翻譯功能
func (ssf *studentGenderFunc) Translation(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(studentGenderTag, fe.Field())
	if err != nil {
		panic(err)
	}

	return t
}

type studentStatusFunc struct {
}

// Validator 提供驗證資料正確性 func
func (ssf *studentStatusFunc) Validator(fl validator.FieldLevel) bool {
	if val, ok := fl.Field().Interface().(int); ok {
		if val == statusDisable || val == statusEnable {
			return true
		}
		return false
	}

	return true
}

// Translations 提供錯誤訊息格式
func (ssf *studentStatusFunc) Translations(ut ut.Translator) error {
	return ut.Add(studentStatusTag, "{0} only Enable(1) or Disable(0)", true)
}

// Translation 提供翻譯功能
func (ssf *studentStatusFunc) Translation(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(studentStatusTag, fe.Field())
	if err != nil {
		panic(err)
	}

	return t
}
