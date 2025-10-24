package security

import "github.com/astaxie/beego/context"

func SecurityHeaders(ctx *context.Context) {
	ctx.Output.Header("Clear-Site-Data", "'cache', 'cookies', 'storage', 'executionContexts'")
	ctx.Output.Header("Cross-Origin-Embedder-Policy", "require-corp")
	ctx.Output.Header("Cross-Origin-Opener-Policy", "same-origin")
	ctx.Output.Header("Cross-Origin-Resource-Policy", "same-origin")
	ctx.Output.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	ctx.Output.Header("Referrer-Policy", "no-referrer")
	ctx.Output.Header("Server", "")
	ctx.Output.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	ctx.Output.Header("X-Content-Type-Options", "nosniff")
	ctx.Output.Header("X-Frame-Options", "DENY")
	ctx.Output.Header("X-Permitted-Cross-Domain-Policies", "none")
}
