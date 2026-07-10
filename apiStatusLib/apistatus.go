package apiStatus

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func Init() {
	beego.Get("/", func(ctx *context.Context) {
		_ = ctx.Output.JSON(map[string]any{"status": "ok"}, true, false)
	})
}
