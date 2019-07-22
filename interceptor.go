package main

import (
  "github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
  "github.com/udistrital/utils_oas/request"
)

func ListenRequest(ctx *context.Context) {
  request.SetHeader(ctx)

}

func InitInterceptor() {
	 beego.InsertFilter("*", beego.BeforeExec, ListenRequest, false)
}
