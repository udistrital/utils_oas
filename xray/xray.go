package xray

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-xray-sdk-go/v2/header"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/aws/aws-xray-sdk-go/v2/xraylog"
)

const (
	segmentKey string = "xray_seg"
	traceIDKey string = "X-Amzn-Trace-Id"
)

var appName = beego.AppConfig.String("appname")
var globalCtx context.Context

func Init() {
	if err := InitXRay(); err != nil {
		logs.Error(err.Error())
	}
}

// Deprecated. Use Init instead. This function is kept for backward compatibility
func InitXRay() error {
	if err := configureXRay(); err != nil {
		return err
	}

	beego.InsertFilter("/:version/*", beego.BeforeExec, beginSegment)
	beego.InsertFilter("/:version/*", beego.AfterExec, EndSegment, false)
	return nil
}

// configureXRay performs the core X-Ray initialization and configuration.
// Returns an error if configuration fails, or nil if X-Ray is not configured or succeeds.
func configureXRay() error {
	if os.Getenv("AWS_XRAY_DAEMON_ADDRESS") == "" {
		return fmt.Errorf("x-ray daemon address not set in environment variable AWS_XRAY_DAEMON_ADDRESS")
	}

	ss, err := xray.NewDefaultStreamingStrategyWithMaxSubsegmentCount(5)
	if err != nil {
		return fmt.Errorf("error creando streaming strategy: %v", err)
	}

	config := xray.Config{StreamingStrategy: ss}
	if err := xray.Configure(config); err != nil {
		return fmt.Errorf("error configurando xray: %v", err)
	}

	xray.SetLogger(xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelInfo))
	logs.Info("X-Ray inicializado correctamente")

	return nil
}

// Función que crea el segmento principal asociado a la petición entrante
// y lo almacena en el contexto de la petición para su posterior uso en subsegmentos

func beginSegment(ctx *beegoCtx.Context) {
	host := ctx.Input.Context.Request.Host
	if strings.HasPrefix(host, "localhost") {
		return
	}

	env := ""
	if strings.HasPrefix(host, "pruebas") {
		env = "_test"
	}

	url := ctx.Input.Scheme() + "://" + host + ctx.Input.Context.Request.URL.String()
	method := ctx.Request.Method
	reqCtx, seg := xray.BeginSegment(ctx.Request.Context(), appName+env)
	seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{Method: method, URL: url},
	}

	traceID := ctx.Request.Header.Values(traceIDKey)
	if traceID != nil {
		h := header.FromString(strings.Trim(traceID[0], "[]"))
		seg.TraceID = h.TraceID
		seg.ParentID = h.ParentID
	}

	reqCtx = context.WithValue(reqCtx, segmentKey, seg)
	ctx.Request = ctx.Request.WithContext(reqCtx)

	globalCtx = reqCtx
}

// Actualiza y cierra el segmento asociado a la petición entrante
func EndSegment(ctx *beegoCtx.Context) {
	seg, ok := ctx.Request.Context().Value(segmentKey).(*xray.Segment)
	if !ok || seg == nil {
		return
	}

	status := ctx.ResponseWriter.Status
	if jsonMap, ok := ctx.Input.GetData("json").(map[string]any); ok {
		if s, ok := jsonMap["Status"].(string); ok {
			if num, err := strconv.Atoi(s); err == nil {
				status = num
			}
		}
	}

	if status == 0 {
		status = http.StatusOK
	}

	xray.HttpCaptureResponse(seg, status)
	seg.Close(nil)
}

// BeginSubsegment starts an X-Ray subsegment for an outgoing HTTP request.
// It sets the trace propagation header on req and returns the updated context and subsegment
func BeginSubsegment(ctx context.Context, req *http.Request) (context.Context, *xray.Segment) {
	if xray.GetSegment(ctx) == nil || req == nil {
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
		// added to show non instrumented services in xray
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

func BeginSegmentSec(req *http.Request) (context.Context, *xray.Segment) {
	return BeginSubsegment(globalCtx, req)
}
