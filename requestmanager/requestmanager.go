package requestmanager

import (
	"encoding/json"
	"fmt"

	"github.com/udistrital/utils_oas/responseformat"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

// FillRequestWithPanic ... unmarshal body request to an interface . panic if some error hapen.
func FillRequestWithPanic(c *beego.Controller, output interface{}) {
	validate = validator.New()
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, output); err != nil {
		logs.Error(err.Error())
		panic(err.Error())
	}
	valErr := validate.Struct(output)

	if valErr != nil {
		var errMess []interface{}
		for _, err := range valErr.(validator.ValidationErrors) {
			errMess = append(errMess, fmt.Sprintf("%s", err))
		}
		logs.Error(errMess)
		responseformat.SetResponseFormat(c, errMess, "", 422)
	}
}
