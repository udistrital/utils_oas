package xray

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-xray-sdk-go/v2/header"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/aws/aws-xray-sdk-go/v2/xraylog"
	"github.com/udistrital/utils_oas/ssm"
)

const traceIDKey string = "X-Amzn-Trace-Id"

var appName = beego.AppConfig.String("appname")
var globalCtx context.Context

func Init() {
	if err := InitXRay(); err != nil {
		logs.Error(err.Error())
	}
}

// Deprecated: Use Init instead.
func InitXRay() error {
	if err := configureXRay(); err != nil {
		return err
	}

	beego.InsertFilter("/:version/*", beego.BeforeExec, beginSegment)
	beego.InsertFilter("/:version/*", beego.AfterExec, EndSegment, false)
	return nil
}

// configureXRay configures the X-Ray SDK with the daemon address stored in
// Parameter Store.
func configureXRay() error {
	parameterStore := beego.AppConfig.String("parameterStore")
	if parameterStore == "" {
		return errors.New("no se puede consultar daemon address: parameterStore no configurado")
	}

	daemonAddr, err := ssm.GetValueFromParameterStore(context.Background(), fmt.Sprintf("/%s/utils/xray/DaemonAddr", parameterStore))
	if err != nil {
		return fmt.Errorf("error consultando daemon address: %w", err)
	}

	ss, err := xray.NewDefaultStreamingStrategyWithMaxSubsegmentCount(5)
	if err != nil {
		return fmt.Errorf("error creando streaming strategy: %w", err)
	}

	config := xray.Config{DaemonAddr: daemonAddr, StreamingStrategy: ss}
	if err := xray.Configure(config); err != nil {
		return fmt.Errorf("error configurando xray: %w", err)
	}

	xray.SetLogger(xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelInfo))
	logs.Info("X-Ray inicializado correctamente")

	return nil
}

// beginSegment is a beego filter that starts the segment for an incoming
// request and stores it in the request context. Loopback requests are not
// traced.
func beginSegment(ctx *beegoCtx.Context) {
	host := ctx.Input.Context.Request.Host
	if isLocalhost(host) {
		return
	}

	env := ""
	if strings.HasPrefix(host, "pruebas") {
		env = "_test"
	}

	reqCtx := startSegment(ctx.Request, appName+env, ctx.Input.Scheme())
	ctx.Request = ctx.Request.WithContext(reqCtx)
	globalCtx = reqCtx
}

// startSegment starts the segment for r, continuing the upstream trace when r
// carries a trace header, and returns a context that holds it.
func startSegment(r *http.Request, name, scheme string) context.Context {
	var (
		ctx context.Context
		seg *xray.Segment
	)
	if traceHeader := r.Header.Get(traceIDKey); traceHeader != "" {
		h := header.FromString(strings.Trim(traceHeader, "[]"))
		ctx, seg = xray.NewSegmentFromHeader(r.Context(), name, r, h)
	} else {
		ctx, seg = xray.BeginSegment(r.Context(), name)
	}

	// The SDK records request data only in xray.Handler, which beego does not use.
	seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: r.Method,
			URL:    scheme + "://" + r.Host + r.URL.RequestURI(),
		},
	}

	return ctx
}

func EndSegment(ctx *beegoCtx.Context) {
	seg := xray.GetSegment(ctx.Request.Context())
	if seg == nil {
		return
	}

	status := http.StatusOK

	if ctx.ResponseWriter.Status != 0 {
		status = ctx.ResponseWriter.Status
	} else if ctx.Output.Status != 0 {
		status = ctx.Output.Status
	}

	xray.HttpCaptureResponse(seg, status)
	seg.Close(nil)
}

// BeginSubsegment starts a subsegment for the outgoing request req and sets
// its trace propagation header. If ctx has no segment, it returns ctx and nil.
func BeginSubsegment(ctx context.Context, req *http.Request) (context.Context, *xray.Segment) {
	if ctx == nil || req == nil || xray.GetSegment(ctx) == nil {
		return ctx, nil
	}

	ctx, subseg := xray.BeginSubsegment(ctx, req.Host)
	subseg.Namespace = "remote"
	subseg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{Method: req.Method, URL: req.URL.String()},
	}
	req.Header.Set(traceIDKey, subseg.DownstreamHeader().String())

	return ctx, subseg
}

// CloseSubsegment records the HTTP response status on subseg and closes it.
func CloseSubsegment(subseg *xray.Segment, resp *http.Response, err error) {
	if subseg == nil {
		return
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	} else if err != nil {
		// Without a response, record a gateway error so the call shows as failed.
		var netErr net.Error
		if (errors.As(err, &netErr) && netErr.Timeout()) || errors.Is(err, context.DeadlineExceeded) {
			statusCode = http.StatusGatewayTimeout
		} else {
			statusCode = http.StatusBadGateway
		}
	}

	xray.HttpCaptureResponse(subseg, statusCode)
	subseg.Close(err)
}

// BeginSegmentSec starts a subsegment for req under the most recently traced
// incoming request. It returns nil if there is none.
func BeginSegmentSec(req *http.Request) *xray.Segment {
	_, seg := BeginSubsegment(globalCtx, req)
	return seg
}

// isLocalhost reports whether host, optionally with a port, is "localhost"
// or a loopback IP.
func isLocalhost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")

	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
