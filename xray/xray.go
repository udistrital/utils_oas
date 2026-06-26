package xray

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-xray-sdk-go/v2/header"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/aws/aws-xray-sdk-go/v2/xraylog"
	"github.com/udistrital/utils_oas/ssm"
)

// xrayLogger routes X-Ray SDK log messages through Beego's logger.
// It intercepts UDP "message too long" errors and replaces them with a
// clear warning, since the service continues to work normally when this
// happens (the trace is dropped but the request is served).
type xrayLogger struct{}

func (l xrayLogger) Log(level xraylog.LogLevel, msg fmt.Stringer) {
	s := msg.String()
	if level == xraylog.LogLevelError && strings.Contains(s, "message too long") {
		logs.Warning("[xray] Segmento demasiado grande para UDP, el trace fue descartado. " +
			"Si este aviso es frecuente, reduzca MaxSubsegmentCount en configureXRay.")
		return
	}
	switch level {
	case xraylog.LogLevelDebug:
		logs.Debug("[xray] " + s)
	case xraylog.LogLevelInfo:
		logs.Info("[xray] " + s)
	case xraylog.LogLevelWarn:
		logs.Warning("[xray] " + s)
	default:
		logs.Error("[xray] " + s)
	}
}

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
	parameterStore := beego.AppConfig.String("parameterStore")
	if parameterStore == "" {
		return fmt.Errorf("no se puede consultar daemon address: %v", errors.New("parameterStore no configurado"))
	}

	daemonAddr, err := ssm.GetValueFromParameterStore(context.Background(), fmt.Sprintf("/%s/utils/xray/DaemonAddr", parameterStore))
	if err != nil {
		return fmt.Errorf("error consultando daemon address: %v", err)
	}

	ss, err := xray.NewDefaultStreamingStrategyWithMaxSubsegmentCount(5)
	if err != nil {
		return fmt.Errorf("error creando streaming strategy: %v", err)
	}

	config := xray.Config{DaemonAddr: daemonAddr, StreamingStrategy: ss}
	if err := xray.Configure(config); err != nil {
		return fmt.Errorf("error configurando xray: %v", err)
	}

	xray.SetLogger(xrayLogger{})
	logs.Info("X-Ray inicializado correctamente")

	return nil
}

// Función que crea el segmento principal asociado a la petición entrante
// y lo almacena en el contexto de la petición para su posterior uso en subsegmentos

// maxURLLen is the maximum number of bytes stored in any URL field of a
// segment or subsegment. Keeping this well under the ~64 KB UDP limit
// ensures that even a single streamed subsegment fits in one datagram.
const maxURLLen = 2048

func truncateURL(u string) string {
	if len(u) > maxURLLen {
		logs.Warning("[xray] URL demasiado larga (%d bytes), será truncada a %d bytes. URL completa: %s", len(u), maxURLLen, u)
		return u[:maxURLLen]
	}
	return u
}

func beginSegment(ctx *beegoCtx.Context) {
	host := ctx.Input.Context.Request.Host
	if strings.HasPrefix(host, "localhost") {
		return
	}

	env := ""
	if strings.HasPrefix(host, "pruebas") {
		env = "_test"
	}

	url := truncateURL(ctx.Input.Scheme() + "://" + host + ctx.Input.Context.Request.URL.String())
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
	if req == nil {
		if ctx == nil {
			return context.Background(), nil
		}
		return ctx, nil
	}

	if ctx == nil {
		return req.Context(), nil
	}

	if xray.GetSegment(ctx) == nil {
		return ctx, nil
	}

	ctx, subseg := xray.BeginSubsegment(ctx, req.Host)
	subseg.Namespace = "remote"
	subseg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{Method: req.Method, URL: truncateURL(req.URL.String())},
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
