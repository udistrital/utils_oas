package apistatus

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func Init() {
	beego.Any("/", func(ctx *context.Context) {
		ctx.Output.JSON(map[string]interface{}{"status": "ok"}, true, true)
	})
}
