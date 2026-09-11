package auditoria

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/cache"
	beego "github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

// failingPutCache serves reads normally but refuses every write.
type failingPutCache struct {
	cache.Cache
}

func (f *failingPutCache) Put(context.Context, string, any, time.Duration) error {
	return errors.New("cache unavailable")
}

// call runs resolveUser, recovering the panic that ctx.Abort raises so tests
// can assert on whether the request was rejected.
func call(ctx *beegoCtx.Context, enforceAuth bool) (aborted bool) {
	defer func() {
		if r := recover(); r != nil {
			aborted = true
		}
	}()

	resolveUser(ctx, enforceAuth)

	return false
}

func TestResolveUserWithoutToken(t *testing.T) {
	cases := []struct {
		name        string
		runMode     string
		enforceAuth bool
		wantAbort   bool
	}{
		{"local development is never rejected", beego.DEV, true, false},
		{"local development while permissive", beego.DEV, false, false},
		{"a deployed service rejects while enforcing", beego.PROD, true, true},
		{"a deployed service passes while permissive", beego.PROD, false, false},
		// Any run mode other than dev is a deployment and must enforce.
		{"an unset run mode rejects while enforcing", "", true, true},
		{"a staging run mode rejects while enforcing", "staging", true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withRunMode(t, tc.runMode)
			stub := setup(t, http.StatusOK, `{"sub":"nobody"}`)
			ctx := newCtx("api.example.com", "")

			if got := call(ctx, tc.enforceAuth); got != tc.wantAbort {
				t.Errorf("aborted = %v, want %v", got, tc.wantAbort)
			}
			if stub.calls != 0 {
				t.Errorf("userinfo called %d times, want 0 without a token", stub.calls)
			}
			if got := ctx.Request.Context().Value(userKey); got != nil {
				t.Errorf("userKey = %v, want nil", got)
			}

			if tc.wantAbort {
				if ctx.Output.Status != http.StatusUnauthorized {
					t.Errorf("status = %d, want 401", ctx.Output.Status)
				}
				if got := ctx.Input.GetData("message"); got != "unauthorized" {
					t.Errorf("message = %v, want unauthorized", got)
				}
			}
		})
	}
}

// The Host header is set by the caller, so it must not influence enforcement:
// a request claiming "localhost" would otherwise skip auth on a deployed
// service.
func TestResolveUserHostHeaderCannotAffectEnforcement(t *testing.T) {
	hosts := []string{
		"api.example.com",
		"localhost",
		"localhost:8080",
		"127.0.0.1:8080",
		"[::1]:8080",
		"localhost.attacker.com",
		"localhostfoo.example.com",
		"",
	}

	for _, host := range hosts {
		t.Run(host, func(t *testing.T) {
			withRunMode(t, beego.PROD)
			setup(t, http.StatusOK, `{"sub":"nobody"}`)

			if !call(newCtx(host, ""), true) {
				t.Errorf("Host %q: not rejected, want 401 regardless of the host claimed", host)
			}
		})
	}
}

// Caching is best effort and happens after the identity is attached, so a cache
// outage degrades to "no caching" rather than dropping a user that was just
// validated.
func TestResolveUserKeepsIdentityWhenTheCacheWriteFails(t *testing.T) {
	stub := setup(t, http.StatusOK, `{"sub":"user-abc"}`)
	c = &failingPutCache{Cache: c}

	ctx := newCtx("api.example.com", "Bearer tok-3")

	if call(ctx, true) {
		t.Fatal("aborted, want the request to pass")
	}
	if stub.calls != 1 {
		t.Errorf("userinfo called %d times, want 1", stub.calls)
	}

	reqCtx := ctx.Request.Context()
	if got := reqCtx.Value(userKey); got != "user-abc" {
		t.Errorf("userKey = %v, want user-abc despite the failed cache write", got)
	}
	if got := reqCtx.Value(authorizationKey); got != "Bearer tok-3" {
		t.Errorf("authorizationKey = %v, want the token", got)
	}
}

func TestResolveUserOnCacheMiss(t *testing.T) {
	// localhost is included deliberately: with a token present it is validated
	// like any other host rather than bypassed.
	for _, host := range []string{"api.example.com", "localhost:8080"} {
		t.Run(host, func(t *testing.T) {
			stub := setup(t, http.StatusOK, `{"sub":"user-abc"}`)
			ctx := newCtx(host, "Bearer tok-1")

			if call(ctx, true) {
				t.Fatal("aborted, want the request to pass")
			}
			if stub.calls != 1 {
				t.Errorf("userinfo called %d times, want 1", stub.calls)
			}
			if stub.gotAuth != "Bearer tok-1" {
				t.Errorf("forwarded Authorization = %q, want %q", stub.gotAuth, "Bearer tok-1")
			}

			reqCtx := ctx.Request.Context()
			if got := reqCtx.Value(userKey); got != "user-abc" {
				t.Errorf("userKey = %v, want user-abc", got)
			}
			// Downstream service calls read the token from here.
			if got := reqCtx.Value(authorizationKey); got != "Bearer tok-1" {
				t.Errorf("authorizationKey = %v, want the token", got)
			}

			cached, err := c.Get(context.Background(), "Bearer tok-1")
			if err != nil || cached != "user-abc" {
				t.Errorf("cache = %v (err %v), want user-abc", cached, err)
			}
		})
	}
}

func TestResolveUserOnCacheHit(t *testing.T) {
	stub := setup(t, http.StatusOK, `{"sub":"should-not-be-used"}`)
	if err := c.Put(context.Background(), "Bearer tok-2", "cached-user", time.Minute); err != nil {
		t.Fatalf("cache put: %v", err)
	}

	ctx := newCtx("api.example.com", "Bearer tok-2")

	if call(ctx, true) {
		t.Fatal("aborted, want the request to pass")
	}
	if stub.calls != 0 {
		t.Errorf("userinfo called %d times, want 0 on a cache hit", stub.calls)
	}

	reqCtx := ctx.Request.Context()
	if got := reqCtx.Value(userKey); got != "cached-user" {
		t.Errorf("userKey = %v, want cached-user", got)
	}
	// Regression guard: the cache-hit path must still carry the token forward.
	if got := reqCtx.Value(authorizationKey); got != "Bearer tok-2" {
		t.Errorf("authorizationKey = %v, want the token", got)
	}
}

func TestResolveUserWhenUpstreamRejectsTheToken(t *testing.T) {
	cases := []struct {
		name        string
		enforceAuth bool
		wantAbort   bool
	}{
		{"enforcing rejects the request", true, true},
		{"permissive lets it through unidentified", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := setup(t, http.StatusUnauthorized, `{}`)
			ctx := newCtx("api.example.com", "Bearer bad")

			if got := call(ctx, tc.enforceAuth); got != tc.wantAbort {
				t.Errorf("aborted = %v, want %v", got, tc.wantAbort)
			}
			if stub.calls != 1 {
				t.Errorf("userinfo called %d times, want 1", stub.calls)
			}
			if got := ctx.Request.Context().Value(userKey); got != nil {
				t.Errorf("userKey = %v, want nil when the token is rejected", got)
			}

			cached, err := c.Get(context.Background(), "Bearer bad")
			if err == nil {
				t.Errorf("cache holds %v, want a rejected token left uncached", cached)
			}
		})
	}
}
