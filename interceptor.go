package auditoria

import (
  "github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
  "github.com/udistrital/utils_oas/request"
)

func InterceptMidRequest(ctx *context.Context) {
  request.SetHeader(ctx.Request.Header["Authorization"][0])

}

func InitInterceptor() {
	 beego.InsertFilter("*", beego.BeforeExec, InterceptMidRequest, false)
}
