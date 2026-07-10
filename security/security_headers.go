package security

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

var securityHeadersMap = map[string]string{
	"Content-Security-Policy":           "default-src 'none'; frame-ancestors 'none'",
	"Cross-Origin-Resource-Policy":      "cross-origin",
	"Permissions-Policy":                "geolocation=(), microphone=(), camera=()",
	"Referrer-Policy":                   "no-referrer",
	"Server":                            "",
	"Strict-Transport-Security":         "max-age=31536000; includeSubDomains; preload",
	"X-XSS-Protection":                  "1; mode=block",
	"X-Content-Type-Options":            "nosniff",
	"X-Frame-Options":                   "DENY",
	"X-Permitted-Cross-Domain-Policies": "none",
}

func SetSecurityHeaders() {
	beego.InsertFilter("*", beego.BeforeExec, securityHeaders)
}

func securityHeaders(ctx *context.Context) {
	for k, v := range securityHeadersMap {
		ctx.Output.Header(k, v)
	}
}
