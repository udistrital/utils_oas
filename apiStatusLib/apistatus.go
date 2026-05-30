package apistatus

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func Init() {
	beego.Get("/", func(ctx *context.Context) {
		_ = ctx.Output.JSON(map[string]any{"status": "ok"}, true, false)
	})
}
