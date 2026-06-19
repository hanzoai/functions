package usagemeter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// When commerce is not configured, the middleware must be a transparent
// pass-through: the handler runs and nothing is gated or charged.
func TestMiddleware_Passthrough_WhenDisabled(t *testing.T) {
	t.Setenv("METERING_DISABLED", "true")

	hit := false
	h := Middleware(zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(200)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/fn/hello", nil))

	if !hit {
		t.Fatal("disabled metering must pass the request through to the handler")
	}
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

// With commerce configured and a positive balance, the middleware gates (allow)
// and records usage for a successful invocation.
func TestMiddleware_GatesAndRecords(t *testing.T) {
	var mu sync.Mutex
	recorded := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { // usage record
			mu.Lock()
			recorded++
			mu.Unlock()
			w.WriteHeader(201)
			_, _ = io.WriteString(w, `{"transactionId":"tx","type":"withdraw"}`)
			return
		}
		_, _ = io.WriteString(w, `{"available":5000}`) // balance
	}))
	defer srv.Close()

	t.Setenv("METERING_DISABLED", "")
	t.Setenv("COMMERCE_URL", srv.URL)
	t.Setenv("COMMERCE_SERVICE_TOKEN", "test-token")
	t.Setenv("COMMERCE_SERVICE_ORG", "hanzo")
	t.Setenv(EnvPricePerCallCents, "3")

	h := Middleware(zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))

	r := httptest.NewRequest(http.MethodGet, "/fn/hello", nil)
	r.Header.Set("X-Org-Id", "hanzo")
	r.Header.Set("X-User-Id", "alice")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200 (balance positive)", rr.Code)
	}

	// Usage record is async; poll briefly.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := recorded
		mu.Unlock()
		if n == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected exactly one usage record for the successful invocation")
}

// With commerce configured and a zero balance, the middleware denies (402) and
// never runs the function.
func TestMiddleware_Denies_WhenNoBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"available":0}`)
	}))
	defer srv.Close()

	t.Setenv("METERING_DISABLED", "")
	t.Setenv("COMMERCE_URL", srv.URL)
	t.Setenv("COMMERCE_SERVICE_TOKEN", "test-token")
	t.Setenv(EnvPricePerCallCents, "3")

	hit := false
	h := Middleware(zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
	}))

	r := httptest.NewRequest(http.MethodGet, "/fn/hello", nil)
	r.Header.Set("X-Org-Id", "hanzo")
	r.Header.Set("X-User-Id", "alice")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)

	if hit {
		t.Fatal("function must NOT run when balance is zero")
	}
	if rr.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d, want 402", rr.Code)
	}
}

// Infra paths bypass metering even when configured.
func TestMiddleware_SkipsInfraPaths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"available":0}`) // would deny if gated
	}))
	defer srv.Close()

	t.Setenv("METERING_DISABLED", "")
	t.Setenv("COMMERCE_URL", srv.URL)
	t.Setenv("COMMERCE_SERVICE_TOKEN", "test-token")
	t.Setenv(EnvPricePerCallCents, "3")

	hit := false
	h := Middleware(zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(200)
	}))

	r := httptest.NewRequest(http.MethodGet, "/router-healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)

	if !hit {
		t.Fatal("/router-healthz must bypass the balance gate")
	}
}
