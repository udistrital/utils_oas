package apistatus

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func Init() {
	beego.Any("/", func(ctx *context.Context) {
		response := map[string]interface{}{"Status": "Ok"}
		ctx.Output.JSON(response, true, true)
	})
}
