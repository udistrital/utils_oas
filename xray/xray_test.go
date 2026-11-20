package xray

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/astaxie/beego/context"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
)

func TestBeginSubsegmentNilContext(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	ctx, seg := BeginSubsegment(nil, req)
	if ctx != nil || seg != nil {
		t.Fatalf("expected nil ctx and segment, got %v, %v", ctx, seg)
	}
}

func TestBeginSubsegmentNoSegmentInContext(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	_, seg := BeginSubsegment(context.Background(), req)
	if seg != nil {
		t.Fatalf("expected nil segment, got %v", seg)
	}
}

func TestBeginSegmentSecWithoutIncomingSegment(t *testing.T) {
	globalCtx = nil
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	if seg := BeginSegmentSec(req); seg != nil {
		t.Fatalf("expected nil segment, got %v", seg)
	}
	// CloseSubsegment must tolerate the nil segment returned above.
	CloseSubsegment(nil, nil, nil)
}

func TestIsLocalhost(t *testing.T) {
	cases := map[string]bool{
		"localhost":             true,
		"localhost:8080":        true,
		"LOCALHOST:8080":        true,
		"127.0.0.1":             true,
		"127.0.0.1:8080":        true,
		"127.1.2.3:80":          true,
		"[::1]:8080":            true,
		"::1":                   true,
		"pruebas.example.com":   false,
		"api.udistrital.edu.co": false,
		"10.0.0.5:8080":         false,
		"localhost.evil.com":    false,
		"":                      false,
	}
	for host, want := range cases {
		if got := isLocalhost(host); got != want {
			t.Errorf("isLocalhost(%q) = %v, want %v", host, got, want)
		}
	}
}

func newBeegoCtx(t *testing.T, host, traceHeader string) *beegoCtx.Context {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "http://"+host+"/v1/arbol/1", nil)
	req.Host = host
	if traceHeader != "" {
		req.Header.Set(traceIDKey, traceHeader)
	}
	ctx := beegoCtx.NewContext()
	ctx.Reset(httptest.NewRecorder(), req)
	return ctx
}

func TestBeginSegmentSkipsLocalhost(t *testing.T) {
	globalCtx = nil
	ctx := newBeegoCtx(t, "127.0.0.1:8080", "")

	beginSegment(ctx)

	if seg := xray.GetSegment(ctx.Request.Context()); seg != nil {
		t.Fatalf("expected no segment for loopback host, got %v", seg)
	}
	if globalCtx != nil {
		t.Fatal("expected globalCtx to stay nil for loopback host")
	}
}

func TestBeginSegmentContinuesIncomingTrace(t *testing.T) {
	const traceID = "1-5759e988-bd862e3fe1be46a994272793"
	const parentID = "53995c3f42cd8ad8"
	ctx := newBeegoCtx(t, "api.example.com", "Root="+traceID+";Parent="+parentID+";Sampled=1")

	beginSegment(ctx)

	seg := xray.GetSegment(ctx.Request.Context())
	if seg == nil {
		t.Fatal("expected a segment in the request context")
	}
	if seg.TraceID != traceID || seg.ParentID != parentID {
		t.Fatalf("got TraceID=%q ParentID=%q, want %q / %q", seg.TraceID, seg.ParentID, traceID, parentID)
	}
	if !seg.Sampled {
		t.Fatal("expected upstream Sampled=1 to be honored")
	}
	if req := seg.GetHTTP().GetRequest(); req.Method != http.MethodGet || req.URL != "http://api.example.com/v1/arbol/1" {
		t.Fatalf("got request data %q %q, want GET http://api.example.com/v1/arbol/1", req.Method, req.URL)
	}

	EndSegment(ctx)
	if seg.GetHTTP().GetResponse().Status != http.StatusOK {
		t.Fatalf("got response status %d, want 200", seg.GetHTTP().GetResponse().Status)
	}
	if seg.InProgress {
		t.Fatal("expected EndSegment to close the segment")
	}
}

func TestBeginSegmentKeepsTraceIDWhenHeaderHasNoRoot(t *testing.T) {
	ctx := newBeegoCtx(t, "api.example.com", "Sampled=1")

	beginSegment(ctx)

	seg := xray.GetSegment(ctx.Request.Context())
	if seg == nil {
		t.Fatal("expected a segment in the request context")
	}
	if seg.TraceID == "" {
		t.Fatal("trace ID was overwritten with an empty value")
	}
	EndSegment(ctx)
}

func TestBeginSegmentHonorsNotSampled(t *testing.T) {
	ctx := newBeegoCtx(t, "api.example.com", "Root=1-5759e988-bd862e3fe1be46a994272793;Sampled=0")

	beginSegment(ctx)

	seg := xray.GetSegment(ctx.Request.Context())
	if seg == nil {
		t.Fatal("expected a segment in the request context")
	}
	if seg.Sampled {
		t.Fatal("expected upstream Sampled=0 to be honored")
	}
	EndSegment(ctx)
}
